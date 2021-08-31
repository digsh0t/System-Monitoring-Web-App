package event

import (
	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/utils"
)

type EventWeb struct {
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
