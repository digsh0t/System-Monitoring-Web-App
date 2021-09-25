package models

import (
	"errors"
	"strconv"

	"github.com/wintltr/login-api/database"
)

type InventoryInfo struct {
	ClientName string
	ClientOS   string
}

type InventoryGroup struct {
	GroupId         int    `json:"groupId"`
	GroupName       string `json:"groupName"`
	SShConnectionId []int  `json:"sshConnectionId"`
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

func GetAllInventoryGroup() ([]InventoryGroup, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT invent_group_id, invent_group_name FROM invent_group`
	selDB, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var inventGroup InventoryGroup
	var inventGroups []InventoryGroup
	for selDB.Next() {
		err = selDB.Scan(&inventGroup.GroupId, &inventGroup.GroupName)
		if err != nil {
			return nil, err
		}
		inventGroups = append(inventGroups, inventGroup)
	}
	return inventGroups, err
}

func InventoryGroupDelete(groupId int) (bool, error) {
	var (
		result bool
		err    error
	)
	result, err = DeleteInventoryGroup(groupId)
	if err != nil {
		return false, errors.New("fail to delete group")
	}

	// Update Inventory
	err = GenerateInventory()
	if err != nil {
		return false, errors.New("fail to update inventory")
	}
	return result, err
}

//Delete SSH Connection Function
func DeleteInventoryGroup(id int) (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM invent_group WHERE invent_group_id = ?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return false, errors.New("no inventory group with this ID exists")
	}
	return true, err
}

func InventoryGroupAddClient(inventGroup InventoryGroup) (bool, error) {
	var (
		result bool
		err    error
	)

	listId := "("
	for index, id := range inventGroup.SShConnectionId {
		if index != 0 {
			listId += ","
		}
		listId += strconv.Itoa(id)
	}
	listId += ")"

	result, err = UpdateGroupId(inventGroup.GroupId, listId)
	if err != nil {
		return false, errors.New("fail to update group id")
	}

	// Update Inventory
	err = GenerateInventory()
	if err != nil {
		return false, errors.New("fail to update inventory")
	}
	return result, err
}

// Update Client Group Id
func UpdateGroupId(groupId int, listId string) (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := "UPDATE ssh_connections SET group_id = ? WHERE sc_connection_id in " + listId
	stmt, err := db.Prepare(query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(groupId)
	if err != nil {
		return false, err
	}
	return true, err
}

func InventoryGroupDeleteClient(inventGroup InventoryGroup) (bool, error) {
	var (
		result bool
		err    error
	)

	listId := "("
	for index, id := range inventGroup.SShConnectionId {
		if index != 0 {
			listId += ","
		}
		listId += strconv.Itoa(id)
	}
	listId += ")"

	result, err = DeleteGroupId(listId)
	if err != nil {
		return false, errors.New("fail to delete group id")
	}

	// Update Inventory
	err = GenerateInventory()
	if err != nil {
		return false, errors.New("fail to update inventory")
	}
	return result, err

}

func DeleteGroupId(listId string) (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := "UPDATE ssh_connections SET group_id = null WHERE sc_connection_id in " + listId
	stmt, err := db.Prepare(query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return false, err
	}
	return true, err

}
