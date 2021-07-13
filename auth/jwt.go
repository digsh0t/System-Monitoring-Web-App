package auth

import (
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/wintltr/login-api/utils"
)

func CreateToken(userId int, username string, role string) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["username"] = username
	claims["role"] = role
	claims["userid"] = userId
	claims["exp"] = time.Now().Add(time.Hour * 12).Unix() // Token expires after 12 hours
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//Activate env file
	utils.EnvInit()
	return token.SignedString([]byte(os.Getenv("SECRET_JWT")))
}
