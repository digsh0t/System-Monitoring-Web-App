package utils

import (
	"log"
	"os/user"
)

// Get Current User interactive
func GetCurrentUser() *user.User {
	user, err := user.Current()
	if err != nil {
		log.Fatal("Failed to load current user!")
	}
	return user
}
