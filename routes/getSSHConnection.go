package routes

import (
	"net/http"

	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

// Get SSh connection from DB
func GetSSHConnection(w http.ResponseWriter, r *http.Request) {

	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT c.sc_username, c.sc_host, c.sc_port, k.sk_key_name 
			  FROM ssh_connections AS c JOIN ssh_keys AS k 
			  ON c.ssh_key_id = k.sk_key_id`
	selDB, err := db.Query(query)
	if err != nil {
		utils.ERROR(w, http.StatusNotFound, "No entry in database!")
	}

	var connectionInfo models.GetSSHConnectionInfo
	var connectionInfos []models.GetSSHConnectionInfo
	for selDB.Next() {
		var username, host, key_name string
		var port int
		err = selDB.Scan(&username, &host, &port, &key_name)
		if err != nil {
			utils.ERROR(w, http.StatusNotFound, "Error while exporting connected clients info!")
		}

		connectionInfo.Sc_username = username
		connectionInfo.Sc_host = host
		connectionInfo.Sc_port = port
		connectionInfo.Sk_key_name = key_name
		connectionInfos = append(connectionInfos, connectionInfo)
	}

	utils.JSON(w, http.StatusOK, connectionInfos)

}
