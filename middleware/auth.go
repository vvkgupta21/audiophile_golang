package middleware

import (
	"audio_phile/model"
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"time"
)

var sampleSecretKey = []byte("GoAudiophileKey")

func AuthMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userToken := r.Header.Get("Authorization")
		checkToken, err := jwt.Parse(userToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("There was an error in parsing token. ")
			}
			return sampleSecretKey, nil
		})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}
		claims, ok := checkToken.Claims.(jwt.MapClaims)
		userData, _ := claims["credential"]
		fmt.Println(userData)
		userInfo, ok := userData.(map[string]interface{})
		fmt.Println(userInfo)

		if ok && checkToken.Valid {
			ctx := context.WithValue(r.Context(), UserContext, userInfo)
			handler.ServeHTTP(w, r.WithContext(ctx))
		}
	})
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
