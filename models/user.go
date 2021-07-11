package models

type User struct {
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
