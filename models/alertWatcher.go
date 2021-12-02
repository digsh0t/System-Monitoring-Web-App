package models

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/wintltr/login-api/database"
)

type WatchInfo struct {
	SSHConnectionId   int    `json:"ssh_connection_id"`
	SSHConnectionName string `json:"ssh_connection_name"`
	WatchList         string `json:"watch_list"`
}

func (watch *WatchInfo) AddNewWatcher() error {
	db := database.ConnectDB()
	defer db.Close()

	query := "INSERT INTO ssh_connection_alert (sca_id, sca_connection_name, sca_alert_pri) VALUES (?,?,?) ON DUPLICATE KEY UPDATE sca_alert_pri = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(watch.SSHConnectionId, watch.SSHConnectionName, watch.WatchList, watch.WatchList)
	return err
}

func RemoveWatcher(sshConnectionId int) error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare(`UPDATE ssh_connection_alert SET sca_alert_pri = "" WHERE sca_id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(sshConnectionId)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return errors.New("this SSH Connection is not in watchlist")
	}
	return err
}

func GetAllWatch() ([]WatchInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT sca_id, sca_connection_name, sca_alert_pri FROM ssh_connection_alert`
	selDB, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var watch WatchInfo
	var watchList []WatchInfo
	for selDB.Next() {
		err = selDB.Scan(&watch.SSHConnectionId, &watch.SSHConnectionName, &watch.WatchList)
		if err != nil {
			return nil, err
		}

		watchList = append(watchList, watch)
	}
	return watchList, err
}

func AlertWatcher() {
	ticker := time.NewTicker(30 * time.Second)
	quit := make(chan struct{})
	func() {
		for {
			select {
			case <-ticker.C:
				err := AllClientAlertLog("/var/log/remotelogs", 30)
				if err != nil {
					if !strings.Contains(err.Error(), "no such file or directory") {
						log.Println(err)
					}
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
