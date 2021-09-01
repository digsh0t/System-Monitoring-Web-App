package routes

import (
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/event"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func hashPassword(password string) string {
	// convert password to byte slice
	var passwordBytes = []byte(password)

	// Load SALT environment variable from .env file
	utils.EnvInit()
	salt := os.Getenv("SALT")
	var saltBytes = []byte(salt)

	passwordBytes = append(passwordBytes, saltBytes...)

	// Create sha-512 hasger
	var sha512Hasher = sha512.New()

	// Write password bytes to the hasher
	sha512Hasher.Write(passwordBytes)

	// Get the SHA-512 hashed password
	var hashedPasswordBytes = sha512Hasher.Sum(nil)

	// Convert hashed password byte array to string
	result := fmt.Sprintf("%x", hashedPasswordBytes)

	return result
}

//Login handler to handle login
func Login(w http.ResponseWriter, r *http.Request) {

	// w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

	var user models.User
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}

	json.Unmarshal(reqBody, &user)

	if !models.CheckInput(user) {
		utils.ERROR(w, 400, "Username and Password length must be greater than 6 characters")
		return
	}

	db := database.ConnectDB()
	defer db.Close()

	hashedPassword := hashPassword(user.Password)
	row := db.QueryRow("SELECT wa_users_id, wa_users_username, wa_users_role FROM wa_users WHERE wa_users_username = ? AND wa_users_password = ?", user.Username, hashedPassword)
	err = row.Scan(&user.UserId, &user.Username, &user.Role)
	var eventStatus string
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, "Wrong Username or Password")
		eventStatus = "failed"
	} else {
		token, err := auth.CreateToken(user.UserId, user.Username, user.Role)
		if err != nil {
			fmt.Println("Fail to create token while login")
		}

		//Return Login Success Authorization Json
		returnJson := simplejson.New()
		returnJson.Set("Authorization", token)
		returnJson.Set("Error", "")
		utils.JSON(w, http.StatusOK, returnJson)
		eventStatus = "successfully"
	}

	// Write Event Web
	description := "User \"" + user.Username + "\" login to web app " + eventStatus
	_, err = event.WriteWebEvent(r, "Login", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write event").Error())
		return
	}

}
