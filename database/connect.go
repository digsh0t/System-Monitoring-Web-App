package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/wintltr/login-api/config"
	"github.com/wintltr/login-api/utils"
)

func ConnectDB() *sql.DB {
	//db, err := sql.Open("mysql", "admin:Anmbmkn123@(capstone-project-db.cjltabe5xft3.us-east-2.rds.amazonaws.com:3306)/CP_Server_Administrator_WA")
	utils.EnvInit()
	db, err := sql.Open("mysql", os.Getenv("DB_STRING"))
	if err != nil {
		log.Fatal(fmt.Println(err))
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(fmt.Println("db.Ping failed: ", err))
	}
	return db
}

func TestConnectionMysqlDB(conf config.ConfigType) (*sql.DB, error) {
	//db, err := sql.Open("mysql", "admin:Anmbmkn123@(capstone-project-db.cjltabe5xft3.us-east-2.rds.amazonaws.com:3306)/CP_Server_Administrator_WA")
	connectionString := conf.MySQL.Username + ":" + conf.MySQL.Password + "@(" + conf.MySQL.Hostname + ")/" + conf.MySQL.DbName
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return db, err
	}
	err = db.Ping()
	return db, err
}

func MigrateDB(db *sql.DB, filepath string) error {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(file), "\n") {
		if line == "" {
			continue
		}
		err = RunQuery(db, strings.Trim(line, "\r\n\t "))
		if err != nil {
			return err
		}
	}
	return err
}

func RunQuery(db *sql.DB, query string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err := db.ExecContext(ctx, query)
	return err
}
