package handler

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	gomail "gopkg.in/gomail.v2"

	"github.com/aitour/scene/auth"
	"github.com/aitour/scene/model"
	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"

	"github.com/oschwald/geoip2-golang"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	tokenProvider = auth.GetDefaultTokenProvider()
	geodb         *geoip2.Reader

	jwtSignMethod  = "HS256"
	jwtHmacSignKey = "tomorrow is saturday"

	isoLangMap = map[string]string{
		"cn": "zh-hans",
		"hk": "zh-hant",
		"tw": "zh-hant",
		"mo": "zh-hant",
		"jp": "ja",
		"gb": "en",
		"us": "en",
	}
)

func init() {
	var err error
	geodb, err = geoip2.Open("assets/GeoLite2-Country.mmdb")
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("load GeoLite2-Country err")
	}
}

func sendAccountActiveMailTo(email string, url string) error {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", "noreply@" + cfg.Http.Domain, "webmaster") // 发件人
	m.SetHeader("To",                                            // 收件人
		m.FormatAddress(email, ""),
	)
	m.SetHeader("Subject", "Account Activation") // 主题

	// 正文
	m.SetBody("text/html", fmt.Sprintf(`Hi there!<br/>
<br/>
Somebody just tried to register for a Aitour account<br/>
using this email address. To complete the registration process,<br/>
just follow this link(copy the url and open in web browser if the lick was not clickable):<br/>
<br/>
<a href="%s">%s</a>
<br/>
If you didn't register for a developer account with Aitour, simply<br/>
ignore this email: no action will be taken.<br/>
<br/>
This link will expire in 48 hours.<br/>
<br/>
Please do not reply this mail<br/>
Take care,<br/>
The Aitour Team`, url, url))

	d := gomail.NewPlainDialer("mail." + cfg.Http.Domain, 25, "postmaster@" + cfg.Http.Domain, "123456") // 发送邮件服务器、端口、发件人账号、发件人密码
	err := d.DialAndSend(m)
	return err
}


//authenticate check middleware
func AuthChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken := c.Request.Header.Get("token")
		if len(authToken) == 0 {
			for _, c := range c.Request.Cookies() {
				if c.Name == "token" {
					authToken = c.Value
					break
				}
			}
		}
		if len(authToken) == 0 {
			authToken = c.Query("token")
		}

		authInfo, err := tokenProvider.GetAuthInfo(authToken)
		if authInfo == nil || err != nil {
			c.SetCookie("redirect", c.Request.RequestURI, 0, "/", "", false, false)
			c.Redirect(http.StatusTemporaryRedirect, "/user/signin?redirect="+c.Request.RequestURI)
			fmt.Fprintf(c.Writer, "auth required")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// The user credentials was found, set user's id to key AuthUserKey in this context, the user's id can be read later using
		// c.MustGet(gin.AuthUserKey).
		c.Set(gin.AuthUserKey, authInfo.User)
	}
}

// gin middleware to set user id in the gin.Context
func AttachUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var lang string
		authToken := c.Request.Header.Get("token")
		if len(authToken) == 0 {
			for _, c := range c.Request.Cookies() {
				if c.Name == "token" {
					authToken = c.Value
				} else if c.Name == "lang" {
					lang = c.Value
				}
			}
		}
		if len(authToken) == 0 {
			authToken = c.Query("token")
		}
		if len(authToken) > 0 {
			//log.Printf("auth token:%s", authToken)
			authInfo, err := tokenProvider.GetAuthInfo(authToken)
			if err != nil {
				log.WithFields(log.Fields{"token": authToken, "err": err}).Debug("get auth info error")
			} else {
				c.Set(gin.AuthUserKey, authInfo.User)
			}
		}

		if len(lang) == 0 {
			lang = c.DefaultQuery("lang", "en")
			c.Set("lang", lang)
		}

		c.Next()
	}
}

func NewCaptacha(c *gin.Context) {
	id := captcha.New()
	c.JSON(http.StatusOK, gin.H{
		"img": fmt.Sprintf("%s.png", id),
	})
}

type RegisterInfo struct {
	Email        string `form:"email"`
	Password     string `form:"password"`
	Language     string `form:"language"`
	VerifyCode   string `form:"vcode"`
	VerifyCodeId string `form:"vcodeid"`
}

func CreateUser(c *gin.Context) {
	var reg RegisterInfo
	if c.ShouldBind(&reg) == nil {
		log.WithFields(log.Fields{"form": reg}).Info("create user")
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bad request parameters",
		})
	}

	if !captcha.VerifyString(reg.VerifyCodeId, reg.VerifyCode) {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"error": "check code verification failed",
			"cv":    captcha.New(),
		})
		return
	}

	user, err := model.CreateUser(reg.Email, reg.Password, reg.Language)
	if err != nil {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"error": err,
			"cv":    captcha.New(),
		})
		return
	}
	log.WithFields(log.Fields{"id": user.Id, "email": user.Email}).Info("user was created")

	tk := jwt.NewWithClaims(jwt.GetSigningMethod(jwtSignMethod),
		jwt.MapClaims{
			"action":    "account.activation",
			"email":     reg.Email,
			"expiresAt": time.Now().Add(48 * time.Hour).UnixNano(),
		})

	activateKey, _ := tk.SignedString([]byte(jwtHmacSignKey))
	activateURL := fmt.Sprintf("https://www." + cfg.Http.Domain + "/user/activate?key=%s", url.QueryEscape(activateKey))
	err = sendAccountActiveMailTo(user.Email, activateURL)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "email": reg.Email}).Warn("send user account activation mail failed")
	}

	c.HTML(http.StatusOK, "register.html", gin.H{
		"regok": true,
	})
	return
}

func ActivateUser(c *gin.Context) {
	key := c.Query("key")

	parsed, err := jwt.Parse(key, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtHmacSignKey), nil
	})
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Warn("jwt parse error on user account activation process")
		c.HTML(http.StatusOK, "register.html", gin.H{
			"activatefail": true,
		})
		return
	}

	claims := parsed.Claims.(jwt.MapClaims)
	action, _ := claims["action"].(string)
	email, _ := claims["email"].(string)
	expiresAt := int64(claims["expiresAt"].(float64))
	if action != "account.activation" || len(email) == 0 || expiresAt == 0 {
		log.WithFields(log.Fields{"action": action, "email": email, "expiresAt": expiresAt}).Warn("invalid account activation token")
		c.HTML(http.StatusOK, "register.html", gin.H{
			"activatefail": true,
		})
		return
	}
	if expiresAt < time.Now().UnixNano() {
		//the activate token expires
		log.WithFields(log.Fields{"expiresAt": expiresAt}).Warn("user activation failed: token expires")
		c.HTML(http.StatusOK, "register.html", gin.H{
			"activatefail": true,
		})
		return
	}

	//log.Printf("activate user:%s", string(email))
	if !model.ActivateUser(string(email)) {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"activatefail": true,
		})
		return
	} else {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"activateok": true,
		})
	}
}

func ChangePwd(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "changepwd.html", gin.H{})
		return
	}

	originPwd := strings.Trim(c.PostForm("originpwd"), " ")
	uid, err := strconv.ParseInt(c.GetString(gin.AuthUserKey), 10, 64)
	if err != nil {
		c.HTML(http.StatusOK, "changepwd.html", gin.H{
			"error": GinTr(c, "invalid user"),
		})
		return
	}
	_, err = model.VerifyUserById(uid, originPwd)
	if err != nil {
		c.HTML(http.StatusOK, "changepwd.html", gin.H{
			"error": GinTr(c, "invalid user"),
		})
		return
	}

	password := strings.Trim(c.PostForm("password"), " ")
	if len(password) == 0 {
		c.HTML(http.StatusOK, "changepwd.html", gin.H{
			"error": GinTr(c, "invalid password"),
		})
		return
	}

	if err = model.SetUserPassword(uid, password); err != nil {
		log.WithError(err).Error("change user password error")
		c.HTML(http.StatusOK, "changepwd.html", gin.H{
			"error": GinTr(c, "set password error"),
		})
		return
	}

	c.HTML(http.StatusOK, "changepwd.html", gin.H{
		"msg": GinTr(c, "change password successfully"),
	})
}

func UserLogin(c *gin.Context) {
	var lang = "en"
	remoteIP := net.ParseIP(c.Request.RemoteAddr[:strings.Index(c.Request.RemoteAddr, ":")])
	if country, err := geodb.Country(remoteIP); err == nil {
		if maplang, ok := isoLangMap[strings.ToLower(country.Country.IsoCode)]; ok {
			lang = maplang
		}
	}

	uid := c.GetString(gin.AuthUserKey)
	if len(uid) > 0 {
		//user already login
		forwardAfterLogin(c)
		return
	}

	c.HTML(http.StatusOK, "signin.html", gin.H{
		"forward": c.Query("forward"),
		"cv":      captcha.New(),
		"lang":    lang,
		"cn":      lang == "zh-hans",
	})
	return
}

func forwardAfterLogin(c *gin.Context) {
	if forward, err := c.Cookie("redirect"); err == nil {
		if len(forward) > 0 {
			c.SetCookie("redirect", "", -1, "/", "", false, false)
			c.Redirect(http.StatusSeeOther, forward)
			return
		}
	}
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func signInUser(c *gin.Context, uid int64) {
	token, _ := tokenProvider.AssignToken(fmt.Sprintf("%d", uid))
	c.SetCookie("token", token, 0, "/", "", false, false)

	if profile, err := model.GetUserProfile(uid); err == nil {
		c.SetCookie("lang", profile.Lang, 0, "/", "", false, false)
	}

	forwardAfterLogin(c)
}

func AuthUser(c *gin.Context) {
	forward := c.PostForm("forwardurl")
	vcode := c.PostForm("vcode")
	vcodeid := c.PostForm("vcodeid")
	email := c.PostForm("email")
	passwrod := c.PostForm("password")
	returnType := c.PostForm("return")

	var lang = "en"
	remoteIP := net.ParseIP(c.Request.RemoteAddr[:strings.Index(c.Request.RemoteAddr, ":")])
	if country, err := geodb.Country(remoteIP); err == nil {
		if maplang, ok := isoLangMap[strings.ToLower(country.Country.IsoCode)]; ok {
			lang = maplang
		}
	}

	if !captcha.VerifyString(vcodeid, vcode) {
		c.HTML(http.StatusOK, "signin.html", gin.H{
			"error": "check code verification failed",
			"lang":  lang,
			"cv":    captcha.New(),
		})
		return
	}

	user, err := model.VerifyUser(email, passwrod)
	if err != nil {
		c.HTML(http.StatusOK, "signin.html", gin.H{
			"error": err,
			"lang":  lang,
			"cv":    captcha.New(),
		})
		return
	}

	token, _ := tokenProvider.AssignToken(fmt.Sprintf("%d", user.Id))
	c.SetCookie("token", token, 0, "/", "", false, false)

	if profile, err := model.GetUserProfile(user.Id); err == nil {
		c.SetCookie("lang", profile.Lang, 0, "/", "", false, false)
	}

	if returnType == "json" {
		//default post back a json contains the token
		c.JSON(http.StatusOK, gin.H{
			"token": token,
		})
		return
	}

	if len(forward) == 0 {
		forward = "/"
	}

	c.Redirect(http.StatusSeeOther, forward)
}

func Logout(c *gin.Context) {
	if token, err := c.Cookie("token"); err == nil {
		//delete the cookie
		c.SetCookie("token", "", -1, "/", "", false, false)
		tokenProvider.RevokeToken(token)
	}
	c.JSON(http.StatusOK, nil)
}

func getUserId(c *gin.Context) int64 {
	var uid int64
	if v, exists := c.Get(gin.AuthUserKey); exists {
		if iv, ok := v.(int64); ok {
			uid = iv
		} else if sv, ok := v.(string); ok {
			uid, _ = strconv.ParseInt(sv, 10, 64)
		}
	}
	return uid
}

func GetUserPhotos(c *gin.Context) {
	uid := getUserId(c)
	photos, err := model.GetUserPhotos(uid)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"uid": uid,
		}).Warn("get user photos error")
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"photos": photos,
	})
}

func GetUserProfile(c *gin.Context) {
	uid := getUserId(c)
	profile, err := model.GetUserProfile(uid)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Warn("get user profile error")
		c.JSON(http.StatusOK, gin.H{
			"error": "internal error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"profile": profile,
	})
}

func SetUserProfile(c *gin.Context) {
	lang := c.PostForm("lang")
	nickName := c.PostForm("nickname")
	avatar := c.PostForm("avatar")

	uid := getUserId(c)
	params := make(map[string]string)

	if len(lang) > 0 {
		params["lang"] = lang
	}
	if len(nickName) > 0 {
		params["nickname"] = nickName
	}
	if len(avatar) > 0 {
		params["avatar"] = avatar
	}
	if len(params) > 0 {
		err := model.SetUserProfile(uid, params)
		if err != nil {
			log.WithFields(log.Fields{
				"uid":    uid,
				"params": params,
				"err":    err,
			}).Warn("set user profile error")

			c.JSON(http.StatusOK, gin.H{
				"error": "internal error",
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{})
}
