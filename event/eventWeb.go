package event

import (
	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/utils"
)

type EventWeb struct {
	EventWebId          int
	EventWebType        string
	EventWebDescription string
	EventWebTimeStamp   string
	EventWebCreatorId   int
}

func (eventWeb *EventWeb) WriteWebEvent() (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO event_web (ev_web_type, ev_web_description, ev_web_timestamp, ev_web_creator_id) VALUES(?,?,?,?)")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	eventWeb.EventWebTimeStamp = utils.GetCurrentDateTime()
	_, err = stmt.Exec(eventWeb.EventWebType, eventWeb.EventWebDescription, eventWeb.EventWebTimeStamp, eventWeb.EventWebCreatorId)
	if err != nil {
		return false, err
	}
	return true, err
}

func GetAllEventWeb() ([]EventWeb, error) {
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
