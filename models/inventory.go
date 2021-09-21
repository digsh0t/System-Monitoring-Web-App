package models

import (
	"errors"

	"github.com/wintltr/login-api/database"
)

type InventoryInfo struct {
	ClientName string
	ClientOS   string
}

type InventoryGroup struct {
	GroupName string `json:"groupName"`
}

func InventoryGroupAdd(inventGroup InventoryGroup) (bool, error) {
	var (
		result bool
		err    error
	)
	result, err = InsertInventoryGroupToDB(inventGroup)
	if err != nil {
		return false, errors.New("fail to insert new group")
	}
	return result, err
}

func InsertInventoryGroupToDB(inventGroup InventoryGroup) (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO invent_group (invent_group_name) VALUES (?)")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(inventGroup.GroupName)
	if err != nil {
		return false, err
	}
	return true, err
}
