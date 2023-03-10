package middleware

import (
	"audio_phile/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type ContextKeys string

const (
	UserContext ContextKeys = "userInfo"
)

func UserContextData(c *gin.Context) (string, error) {
	user, exists := c.Get(string(UserContext))
	if !exists {
		return "", fmt.Errorf("user context is not set")
	}
	userData, ok := user.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("failed to parse user context data")
	}
	role, ok := userData["role"].(string)
	if !ok {
		return "", fmt.Errorf("user role is not a string")
	}
	return role, nil
}

func AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, err := UserContextData(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Println(err)
			return
		}
		fmt.Println(role)
		if model.Role(role) == model.RoleAdmin {
			ctx.Next()
			return
		}
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "user does not have the necessary permissions"})
	}
}

func UserMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, err := UserContextData(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Println(role)
		if model.Role(role) == model.RoleUser {
			ctx.Next()
			return
		}
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "user does not have the necessary permissions"})
	}
}
