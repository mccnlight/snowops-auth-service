package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nurpe/snowops-auth/internal/adapter/token"
)

const TokenClaimsKey = "tokenClaims"

func Auth(tokenManager *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header missing"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		claims, err := tokenManager.ParseAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(TokenClaimsKey, claims)
		c.Next()
	}
}

func GetClaims(c *gin.Context) (*token.Claims, bool) {
	value, ok := c.Get(TokenClaimsKey)
	if !ok {
		return nil, false
	}

	claims, ok := value.(*token.Claims)
	if !ok {
		return nil, false
	}

	return claims, true
}
