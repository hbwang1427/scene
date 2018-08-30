package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/aitour/scene/model"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

var (
	qqClientId = ""
	authStates = make(map[string]int64)
)

func AuthQQ(c *gin.Context) {
	// qqClientId := ""
	// redirectUrl := ""
	// responseType := "code"
	// state := ""
	// url := `https://graph.qq.com/oauth2.0/show?which=Login&display=pc&client_id=101284669&redirect_uri=https%3A%2F%2Fgitee.com%2Fauth%2Fqq_connect%2Fcallback&response_type=code&state=4188dd9902cf5bb2c8279a0557324b1860db3f5d000ece56`
}

func AuthWechat(c *gin.Context) {

}

func AuthWeibo(c *gin.Context) {

}

func bindOpenId(c *gin.Context, platform string, email string, openid string, name string, picture string, locale string) {
	//检查数据库是否存在Email对应的账户
	user, err := model.VerifyUserByOpenId(platform, openid)
	// if err != nil {
	// 	log.WithFields(log.Fields{"err": err}).Error("Verify user by openid error")
	// 	c.HTML(http.StatusOK, "signin.html", gin.H{
	// 		"error": err,
	// 	})
	// 	return
	// }

	if user == nil {
		//创建并绑定账户
		if user, err = model.BindOpenId(platform, email, openid); err != nil {
			log.WithFields(log.Fields{"err": err, "platform": platform, "email": email, "openid": openid}).Error("bind user error")
			c.HTML(http.StatusOK, "signin.html", gin.H{
				"error": err,
			})
			return
		}

		//设置avartar
		log.WithFields(log.Fields{"uid": user.Id, "avatar": picture}).Info("set user profile")
		err = model.SetUserProfile(user.Id, map[string]string{
			"nickname": name,
			"avatar":   picture,
			"lang":     locale,
		})
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("set user profile error")
			c.HTML(http.StatusOK, "signin.html", gin.H{
				"error": err,
			})
			return
		}
	}

	if user != nil {
		signInUser(c, user.Id)
	}
}

func bindGoogle(c *gin.Context) {
	fbConfig := oauth2.Config{
		ClientID:     "216679058012-21gqlhp3eh1qp4mmvtmag7298h991udb.apps.googleusercontent.com",
		ClientSecret: "OWwBt-b2XVz5qLDeGDaKEsVJ",
		RedirectURL:  "https://aitour.ml/openid/google",
		Scopes: []string{
			"https://www.googleapis.com/auth/plus.me",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	state := c.Query("state")
	if len(state) == 0 {
		state = fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Int31())
		authStates[state] = time.Now().UnixNano()
		authCodeURL := fbConfig.AuthCodeURL(state)
		c.Redirect(http.StatusTemporaryRedirect, authCodeURL)
		return
	}

	if _, ok := authStates[state]; !ok {
		c.Writer.Write([]byte("invalid state"))
	}
	delete(authStates, state)

	code := c.Query("code")
	token, err := fbConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Warn("exchange token error")
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	var userInfo struct {
		ID      string
		Name    string
		Email   string
		Picture string
		Gender  string
		Locale  string
	}
	content, _ := ioutil.ReadAll(response.Body)
	if err := json.Unmarshal(content, &userInfo); err != nil {
		log.WithFields(log.Fields{"err": err, "content": content}).Error("decode facebook user info error")
		c.HTML(http.StatusOK, "signin.html", gin.H{
			"error": err,
		})
		return
	}
	bindOpenId(c, "fb", userInfo.Email, userInfo.ID, userInfo.Name, userInfo.Picture, userInfo.Locale)
	// defer response.Body.Close()
	// contents, err := ioutil.ReadAll(response.Body)
	// fmt.Fprintf(c.Writer, "Content: %s\n", contents)
}

func bindFaceBook(c *gin.Context) {
	fbConfig := oauth2.Config{
		ClientID:     "186885222135849",
		ClientSecret: "8aac6edf699e196fd9259c86d07d2414",
		RedirectURL:  "https://aitour.ml/openid/facebook",
		Scopes:       []string{"public_profile", "email"}, //, "user_hometown", "user_birthday", "user_gender"
		Endpoint:     facebook.Endpoint,
	}

	state := c.Query("state")
	if len(state) == 0 {
		state = fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Int31())
		authStates[state] = time.Now().UnixNano()
		authCodeUrl := fbConfig.AuthCodeURL(state)
		c.Redirect(http.StatusTemporaryRedirect, authCodeUrl)
		return
	}

	if _, ok := authStates[state]; !ok {
		fmt.Fprintf(c.Writer, "invalid state:%s", state)
		return
	}
	delete(authStates, state)

	code := c.Query("code")
	token, err := fbConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Warn("exchange token error")
	}

	query := &url.Values{}
	query.Set("access_token", token.AccessToken)
	query.Set("fields", "id,name,email,picture.type(large)")
	query.Set("method", "get")
	query.Set("sdk", "joey")
	query.Set("suppress_http_code", "1")
	response, err := http.Get("https://graph.facebook.com/v2.8/me?" + query.Encode())

	var userInfo struct {
		ID      string
		Name    string
		Email   string
		Picture struct {
			Data struct {
				Width  int
				Height int
				Url    string
			}
		}
	}

	content, _ := ioutil.ReadAll(response.Body)
	if err := json.Unmarshal(content, &userInfo); err != nil {
		log.WithFields(log.Fields{"err": err, "content": content}).Error("decode facebook user info error")
		c.HTML(http.StatusOK, "signin.html", gin.H{
			"error": err,
		})
		return
	}
	bindOpenId(c, "fb", userInfo.Email, userInfo.ID, userInfo.Name, userInfo.Picture.Data.Url, "")
}

func SetupThirdPartyAuthHandlers(r *gin.Engine) {
	r.GET("/openid/qq", AuthQQ)
	r.GET("/openid/weixin", AuthWechat)
	r.GET("/openid/weibo", AuthWeibo)

	r.GET("/openid/google", bindGoogle)
	r.GET("/openid/facebook", bindFaceBook)
}
