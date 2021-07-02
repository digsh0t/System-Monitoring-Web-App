package utils

import (
	"log"

	"github.com/joho/godotenv"
)

func Init() {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
