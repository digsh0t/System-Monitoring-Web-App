package utils

import (
	"log"

	"github.com/joho/godotenv"
)

//Load .env file
func EnvInit() {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file, please create one")
	}
}
