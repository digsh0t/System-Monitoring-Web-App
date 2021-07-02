package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDB() *sql.DB {
	db, err := sql.Open("mysql", "root:Anmbmnk123@(127.0.0.1:3306)/CP_Server_Administrator_WA")
	if err != nil {
		log.Fatal(fmt.Println(err))
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(fmt.Println("db.Ping failed: ", err))
	}
	return db
}
