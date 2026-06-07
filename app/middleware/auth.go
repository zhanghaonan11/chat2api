package middleware

import (
	"strings"

	"chat2api/app/conf"

	"github.com/gin-gonic/gin"
)

func V1Auth(c *gin.Context) {
	authToken := c.Request.Header.Get("Authorization")
	localToken := strings.TrimSpace(strings.TrimPrefix(authToken, "Bearer "))
	if strings.HasPrefix(localToken, "at-") {
		c.Next()
		return
	}
	if strings.HasPrefix(authToken, "Bearer eyJhbGciOiJSUzI1NiI") {
		c.Next()
		return
	}
	tokens := conf.GetAuthAccessTokens()
	if len(tokens) == 0 {
		authError(c, "No local API keys are configured")
		return
	}
	if authToken == "" {
		authError(c, "You didn't provide an API key. You need to provide your API key in an Authorization header using Bearer auth (i.e. Authorization: Bearer YOUR_KEY)")
		return
	}
	if !containsAuthToken(tokens, localToken) {
		authError(c, "Incorrect API key provided: sk-4yNZz***************************************6mjw.")
		return
	}
	c.Next()
}

func containsAuthToken(tokens []string, token string) bool {
	for _, item := range tokens {
		if item == token {
			return true
		}
	}
	return false
}

func authError(c *gin.Context, message string) {
	c.AbortWithStatusJSON(401, gin.H{
		"detail": gin.H{
			"code":  401,
			"msg":   message,
			"error": nil,
		},
	})
}
