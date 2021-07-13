package routes

import (
	"net/http"

	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

// Get SSh connection from DB
func GetAllSSHConnection(w http.ResponseWriter, r *http.Request) {

	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT sc_connection_id, sc_username, sc_host, sc_port, creator_id, ssh_key_id 
			  FROM ssh_connections`
	selDB, err := db.Query(query)
	if err != nil {
		utils.ERROR(w, http.StatusNotFound, "No entry in database!")
	}

	var connectionInfo models.SshConnectionInfo
	var connectionInfos []models.SshConnectionInfo
	for selDB.Next() {
		err = selDB.Scan(&connectionInfo.SSHConnectionId, &connectionInfo.UserSSH, &connectionInfo.HostSSH, &connectionInfo.PortSSH, &connectionInfo.CreatorId, &connectionInfo.SSHKeyId)
		if err != nil {
			utils.ERROR(w, http.StatusNotFound, "Error while exporting connected clients info!")
		}

		connectionInfos = append(connectionInfos, connectionInfo)
	}

	utils.JSON(w, http.StatusOK, connectionInfos)

}
