package routes

import (
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
	row := db.QueryRow("SELECT wa_users_username, wa_users_role FROM wa_users WHERE wa_users_username = ? AND wa_users_password = ?", user.Username, user.Password)
	err = row.Scan(&user.Username, &user.Role)
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, "Wrong Username or Password")
	} else {
		token, err := auth.Create(user.Username, user.Role)
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
