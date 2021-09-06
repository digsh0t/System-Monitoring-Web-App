package event

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

type EventWeb struct {
	EventWebId          int
	EventWebType        string
	EventWebDescription string
	EventWebTimeStamp   string
	EventWebCreatorId   int
}

func WriteWebEvent(r *http.Request, eventType string, description string) (bool, error) {
	var id int
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO event_web (ev_web_type, ev_web_description, ev_web_timestamp, ev_web_creator_id) VALUES(?,?,?,?)")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	// Get User Id
	if eventType != "Login" {
		id, err = auth.ExtractUserId(r)
		if err != nil {
			return false, err
		}
	} else {
		pattern := "\"(.+)\""
		r, _ := regexp.Compile(pattern)
		submatch := r.FindStringSubmatch(description)
		username := submatch[1]
		id, err = models.GetIdFromUsername(username)
		if err != nil {
			return false, err
		}
	}

	timestamp := utils.GetCurrentDateTime()
	_, err = stmt.Exec(eventType, description, timestamp, id)
	if err != nil {
		return false, err
	}
	return true, err
}

func GetAllEventWeb(r *http.Request) ([]EventWeb, error) {

	var eventWebList []EventWeb
	tokenData, err := auth.RetrieveTokenData(r)
	if err != nil {
		return eventWebList, errors.New("Failed to retrieve user token")
	}
	if tokenData.Role == "admin" {
		eventWebList, err = GetAllEventWebFromDB()
	} else {
		eventWebList, err = GetEventWebFromUserID(tokenData.Userid)
	}
	if err != nil {
		return eventWebList, errors.New("Failed to get all event web")
	}
	return eventWebList, err

}

func GetAllEventWebFromDB() ([]EventWeb, error) {
	db := database.ConnectDB()
	defer db.Close()

	var eventWebList []EventWeb
	selDB, err := db.Query("SELECT * FROM event_web")
	if err != nil {
		return eventWebList, err
	}

	var eventWeb EventWeb
	for selDB.Next() {
		var ev_id, creatorId int
		var ev_type, ev_description, ev_timestamp string

		err = selDB.Scan(&ev_id, &ev_type, &ev_description, &ev_timestamp, &creatorId)
		if err != nil {
			return eventWebList, err
		}
		eventWeb = EventWeb{
			EventWebId:          ev_id,
			EventWebType:        ev_type,
			EventWebDescription: ev_description,
			EventWebTimeStamp:   ev_timestamp,
			EventWebCreatorId:   creatorId,
		}
		eventWebList = append(eventWebList, eventWeb)
	}

	return eventWebList, err

}

func GetEventWebFromUserID(userId int) ([]EventWeb, error) {
	db := database.ConnectDB()
	defer db.Close()

	var eventWebList []EventWeb
	selDB, err := db.Query("SELECT * FROM event_web WHERE ev_web_creator_id = ?", userId)
	if err != nil {
		return eventWebList, err
	}

	var eventWeb EventWeb
	for selDB.Next() {
		var ev_id, creatorId int
		var ev_type, ev_description, ev_timestamp string

		err = selDB.Scan(&ev_id, &ev_type, &ev_description, &ev_timestamp, &creatorId)
		if err != nil {
			return eventWebList, err
		}
		eventWeb = EventWeb{
			EventWebId:          ev_id,
			EventWebType:        ev_type,
			EventWebDescription: ev_description,
			EventWebTimeStamp:   ev_timestamp,
			EventWebCreatorId:   creatorId,
		}
		eventWebList = append(eventWebList, eventWeb)
	}

	return eventWebList, err

}

func DeleteAllEventWeb() (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM event_web")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	res, err := stmt.Exec()
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return false, err
	}
	return true, err
}
