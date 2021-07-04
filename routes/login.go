package routes

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func hashPassword(password string) string {
	// convert password to byte slice
	var passwordBytes = []byte(password)

	// Create sha-512 hasger
	var sha512Hasher = sha512.New()

	// Write password bytes to the hasher
	sha512Hasher.Write(passwordBytes)

	// Get the SHA-512 hashed password
	var hashedPasswordBytes = sha512Hasher.Sum(nil)

	// Convert the hashed password to a base64 encoded string
	var base64EncodedPasswordHash = base64.URLEncoding.EncodeToString(hashedPasswordBytes)

	return base64EncodedPasswordHash
}

//Login handler to handle login
func Login(w http.ResponseWriter, r *http.Request) {

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
	row := db.QueryRow("SELECT wa_users_username, wa_users_role FROM wa_users WHERE wa_users_username = ? AND wa_users_hashedPassword = ?", user.Username, hashedPassword)
	err = row.Scan(&user.Username, &user.Role)
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, "Wrong Username or Password")
	} else {
		token, err := auth.CreateToken(user.Username, user.Role)
		if err != nil {
			fmt.Println("Fail to create token while login")
		}

		//Return Login Success Authorization Json

		returnJson := simplejson.New()
		returnJson.Set("Authorization", token)
		returnJson.Set("Error", "")
		utils.JSON(w, http.StatusOK, returnJson)
	}

}