package handler

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aitour/scene/auth"
	"github.com/aitour/scene/model"
	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

var (
	tokenProvider auth.TokenProvider
)

func init() {
	var err error
	tokenProvider, err = auth.CreateTokenProvider("jwt", map[string]interface{}{
		"key":      "hmacsecretkey",
		"tokenTTL": 30 * time.Minute,
		"tokenLen": 16,
	})
	if err != nil {
		log.Fatalln(err)
	}
}

func sendAccountActiveMailTo(email string, url string) error {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", "83527338@qq.com", "webmaster@aitour.top") // 发件人
	m.SetHeader("To",                                                     // 收件人
		m.FormatAddress(email, ""),
	)
	m.SetHeader("Subject", "Account Activation")                                                // 主题
	m.SetBody("text/html", "Hello <a href = \"http://blog.csdn.net/liang19890820\">一去丶二三里</a>") // 正文

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
This link will expire in 24 hours.<br/>
<br/>
Take care,<br/>
The Aitour Team`, url, url))

	d := gomail.NewPlainDialer("smtp.qq.com", 465, "83527338@qq.com", "nddutxiwjdfzcaij") // 发送邮件服务器、端口、发件人账号、发件人密码
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
			c.Redirect(http.StatusTemporaryRedirect, "/user/signin?redirect="+c.Request.RequestURI)
			fmt.Fprintf(c.Writer, "access forbiden")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// The user credentials was found, set user's id to key AuthUserKey in this context, the user's id can be read later using
		// c.MustGet(gin.AuthUserKey).
		c.Set(gin.AuthUserKey, authInfo.UserId)
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
	VerifyCode   string `form:"vcode"`
	VerifyCodeId string `form:"vcodeid"`
}

func CreateUser(c *gin.Context) {
	var reg RegisterInfo
	if c.ShouldBind(&reg) == nil {
		log.Printf("create user: %v", reg)
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

	user, err := model.CreateUser(reg.Email, reg.Password)
	if err != nil {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"error": err,
			"cv":    captcha.New(),
		})
		return
	}
	log.Printf("user %v created", user)

	activateUrl := fmt.Sprintf("http://localhost:8081/user/activate?key=%s_%s",
		base64.StdEncoding.EncodeToString([]byte(user.Email)), user.ActivateKey)
	log.Printf("activate url:%s", activateUrl)
	err = sendAccountActiveMailTo(user.Email, activateUrl)
	if err != nil {
		log.Printf("send user account activation mail failed：%v", err)
	} else {
		log.Printf("send user account activation mail ok")
	}

	c.HTML(http.StatusOK, "register.html", gin.H{
		"regok": true,
	})
	return
}

func ActivateUser(c *gin.Context) {
	key := c.Query("key")
	parts := strings.Split(key, "_")
	var email []byte
	var err error

	if len(parts) != 2 {
		goto ACTIVATE_FAIL
	}
	if email, err = base64.StdEncoding.DecodeString(parts[0]); err != nil {
		goto ACTIVATE_FAIL
	}

	log.Printf("activate user:%s %s", string(email), parts[1])
	if !model.ActivateUser(string(email), parts[1]) {
		goto ACTIVATE_FAIL
	} else {
		goto ACTIVATE_OK
	}

ACTIVATE_FAIL:
	c.HTML(http.StatusOK, "register.html", gin.H{
		"activatefail": true,
	})
	return
ACTIVATE_OK:
	c.HTML(http.StatusOK, "register.html", gin.H{
		"activateok": true,
	})
}

func UserLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "signin.html", gin.H{
		"forward": c.Query("forward"),
		"cv":      captcha.New(),
	})
	return
}

func AuthUser(c *gin.Context) {
	forward := c.PostForm("forwardurl")
	vcode := c.PostForm("vcode")
	vcodeid := c.PostForm("vcodeid")
	email := c.PostForm("email")
	passwrod := c.PostForm("password")
	if !captcha.VerifyString(vcodeid, vcode) {
		c.HTML(http.StatusOK, "signin.html", gin.H{
			"error": "check code verification failed",
			"cv":    captcha.New(),
		})
		return
	}

	user, err := model.VerifyUser(email, passwrod)
	if err != nil {
		c.HTML(http.StatusOK, "signin.html", gin.H{
			"error": err,
			"cv":    captcha.New(),
		})
		return
	}

	token, _ := tokenProvider.AssignToken(user.Id)
	c.SetCookie("token", token, 0, "/", "", false, false)

	if len(forward) > 0 {
		log.Printf("forwart to %s", forward)
		c.Redirect(http.StatusSeeOther, forward)
		return
	}

	//default post back a json contains the token
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func Logout(c *gin.Context) {
	if token, err := c.Cookie("token"); err == nil {
		//delete the cookie
		c.SetCookie("token", "", -1, "/", "", false, false)
		tokenProvider.RevokeToken(token)
	}
	c.JSON(http.StatusOK, nil)
}
