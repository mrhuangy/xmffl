package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"fpxxl/minigame-api/internal/repository"
)

func Auth(repo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			token = c.GetHeader("X-Openid")
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		player, err := repo.PlayerByOpenID(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("player", player)
		c.Next()
	}
}

func bearerToken(value string) string {
	if value == "" {
		return ""
	}
	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
