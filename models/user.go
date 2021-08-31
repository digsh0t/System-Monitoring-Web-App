package models

import (
	"errors"

	"github.com/wintltr/login-api/database"
)

type User struct {
	UserId   int    `json:"userid"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

//Check if username or password is legit
func CheckInput(user User) bool {
	if len(user.Username) < 6 || len(user.Password) < 6 {
		return false
	} else {
		return true
	}
}

func GetIdFromUsername(username string) (int, error) {
	db := database.ConnectDB()
	defer db.Close()

	var id int
	row := db.QueryRow("SELECT wa_users_id FROM wa_users WHERE wa_users_username = ?", username)
	err := row.Scan(&id)
	if row == nil {
		return id, errors.New("ssh connection doesn't exist")
	}

	return id, err
}
