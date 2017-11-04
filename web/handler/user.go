package handler

import (
	"fmt"
	"net/http"

	"github.com/aitour/scene/auth"
	"github.com/gin-gonic/gin"
)

//authenticate check middleware
func AuthChecker(tokenProvider auth.TokenProvider) gin.HandlerFunc {
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
			fmt.Fprintf(c.Writer, "access forbiden")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// The user credentials was found, set user's id to key AuthUserKey in this context, the user's id can be read later using
		// c.MustGet(gin.AuthUserKey).
		c.Set(gin.AuthUserKey, authInfo.UserName)
	}
}

func AuthUser(c *gin.Context) {

}

func Logout(c *gin.Context) {

}
