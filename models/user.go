package models

import (
	"crypto/sha512"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/utils"
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

func GetUsernameFromId(id int) (string, error) {
	db := database.ConnectDB()
	defer db.Close()

	var username string
	row := db.QueryRow("SELECT wa_users_username FROM wa_users WHERE wa_users_id = ?", id)
	err := row.Scan(&username)
	if row == nil {
		return username, errors.New("web app user doesn't exist")
	}

	return username, err
}

func AddWebAppUser(user User) (bool, error) {

	result := CheckInput(user)
	if !result {
		return false, errors.New("username and password must be equal or greater than 6")
	}

	result, err := CheckUserNameExist(user.Username)
	if !result && err != nil {
		return false, errors.New("failed to check user from db")
	}
	if result && err == nil {
		return false, errors.New("username existed, please to another username")
	}

	err = InsertUserToDB(user)
	if err != nil {
		return false, errors.New("failed to create new user")
	}
	return true, err

}

func DeleteWepAppUser(waUserId int) (bool, error) {
	err := DeleteUserFromDB(waUserId)
	if err != nil {
		return false, err
	}
	return true, err
}

func UpdateWepAppUser(user User) (bool, error) {

	result := CheckInput(user)
	if !result {
		return false, errors.New("username and password must be equal or greater than 6")
	}

	beforeUsername, err := GetUsernameFromId(user.UserId)
	if err != nil {
		return false, errors.New("failed to update user")
	}
	if user.Username != beforeUsername {
		result, err := CheckUserNameExist(user.Username)
		if !result && err != nil {
			return false, errors.New("failed to check user from db")
		}
		if result && err == nil {
			return false, errors.New("username existed, please to another username")
		}
	}

	err = UpdateUserToDB(user)
	if err != nil {
		return false, err
	}
	return true, err
}

func ListAllWepAppUser() ([]User, error) {
	var waUserList []User
	waUserList, err := GetAllUserFromDB()
	if err != nil {
		return waUserList, err
	}
	return waUserList, err
}

func ListWepAppUser(waUserId int) (User, error) {
	var waUser User
	waUser, err := GetUserByIdFromDB(waUserId)
	if err != nil {
		if err == sql.ErrNoRows {
			return waUser, errors.New("No user matched with this Id")
		}
		return waUser, errors.New("Failed to list web app user")
	}
	return waUser, err
}

// Insert Wep App User
func InsertUserToDB(user User) error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO wa_users (wa_users_username, wa_users_password, wa_users_role) VALUES (?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	hashedPassword := HashPassword(user.Password)
	_, err = stmt.Exec(user.Username, hashedPassword, user.Role)
	if err != nil {
		return err
	}
	return err
}

//Delete Wep App User
func DeleteUserFromDB(id int) error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM wa_users WHERE wa_users_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return errors.New("no web app user with this ID exists")
	}
	return err
}

// Update Wep App User
func UpdateUserToDB(user User) error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("UPDATE wa_users SET wa_users_username = ?, wa_users_password = ? , wa_users_role = ? WHERE wa_users_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	hashedPassword := HashPassword(user.Password)
	_, err = stmt.Exec(user.Username, hashedPassword, user.Role, user.UserId)
	if err != nil {
		return err
	}
	return err
}

// List All Wep App User
func GetAllUserFromDB() ([]User, error) {
	db := database.ConnectDB()
	defer db.Close()

	var waUserList []User
	selDB, err := db.Query("SELECT wa_users_id, wa_users_username, wa_users_role FROM wa_users")
	if err != nil {
		return waUserList, err
	}

	var user User
	for selDB.Next() {
		var id int
		var username, role string

		err = selDB.Scan(&id, &username, &role)
		if err != nil {
			return waUserList, err
		}
		user = User{
			UserId:   id,
			Username: username,
			Role:     role,
		}
		waUserList = append(waUserList, user)
	}

	return waUserList, err

}

// List Wep App User By Id
func GetUserByIdFromDB(waUserId int) (User, error) {
	db := database.ConnectDB()
	defer db.Close()

	var waUser User
	row := db.QueryRow("SELECT wa_users_id, wa_users_username, wa_users_role FROM wa_users WHERE wa_users_id = ?", waUserId)
	err := row.Scan(&waUser.UserId, &waUser.Username, &waUser.Role)
	if err != nil {
		return waUser, err
	}

	return waUser, err

}

func CheckUserNameExist(username string) (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	err := db.QueryRow("SELECT wa_users_username FROM wa_users WHERE wa_users_username = ?", username).Scan(&username)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}

		return false, nil
	}

	return true, nil
}

func HashPassword(password string) string {
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
