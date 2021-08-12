package auth

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/wintltr/login-api/utils"
)

type TokenData struct {
	Username string
	Role     string
	Userid   int
	Exp      uint64
}

func CreateToken(userId int, username string, role string) (string, error) {
	claims := jwt.MapClaims{}
	claims["username"] = username
	claims["role"] = role
	claims["userid"] = userId
	claims["exp"] = time.Now().Add(time.Hour * 12).Unix() // Token expires after 12 hours
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//Activate env file
	utils.EnvInit()
	return token.SignedString([]byte(os.Getenv("SECRET_JWT")))
}

func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET_JWT")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

//Check if Token has expired
func TokenValid(r *http.Request) error {
	token, err := VerifyToken(r)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}
	return nil
}

//Extract UserId from Token
func ExtractUserId(r *http.Request) (int, error) {
	tokenData, err := ExtractTokenMetadata(r)
	if err != nil {
		return -1, err
	}
	return tokenData.Userid, err
}

func ExtractTokenMetadata(r *http.Request) (*TokenData, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		username, ok := claims["username"].(string)
		if !ok {
			return nil, err
		}
		role, ok := claims["role"].(string)
		if !ok {
			return nil, err
		}
		userId, err := strconv.Atoi(fmt.Sprintf("%.f", claims["userid"]))
		if err != nil {
			return nil, err
		}
		exp, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["exp"]), 10, 64)
		if err != nil {
			return nil, err
		}
		return &TokenData{
			Username: username,
			Role:     role,
			Userid:   userId,
			Exp:      exp,
		}, nil
	}
	return nil, err
}

func CheckAuth(r *http.Request, authorizedRoles []string) (bool, error) {
	tokenData, err := ExtractTokenMetadata(r)
	if err != nil {
		return false, err
	}
	if utils.FindInStringArray(authorizedRoles, tokenData.Role) {
		return true, err
	} else {
		return false, err
	}
}
