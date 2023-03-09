package middleware

import (
	"audio_phile/model"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"time"
)

var sampleSecretKey = []byte("GoAudiophileKey")

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userToken := ctx.GetHeader("Authorization")
		checkToken, err := jwt.Parse(userToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("There was an error in parsing token. ")
			}
			return sampleSecretKey, nil
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Println(err)
			return
		}
		claims, ok := checkToken.Claims.(jwt.MapClaims)
		userData, _ := claims["credential"]
		fmt.Println(userData)
		userInfo, ok := userData.(map[string]interface{})
		fmt.Println(userInfo)

		if ok && checkToken.Valid {
			ctx.Set(string(UserContext), userInfo)
			ctx.Next()
		}
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
}

// GenerateJWT is used to generate the JWT token
func GenerateJWT(userId string, role model.Role) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["credential"] = model.UserCredential{
		Id:    userId,
		Roles: role,
	}
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()
	tokenString, err := token.SignedString(sampleSecretKey)
	if err != nil {
		logrus.Errorf("something Went Wrong: %s", err.Error())
		return "", err
	}
	return tokenString, nil
}
