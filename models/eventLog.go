package models

import (
	"time"

	"github.com/wintltr/login-api/database"
)

type Event struct {
	EventId     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	TimeStampt  time.Time `json:"created"`
	CreatorId   int       `json:"creator_id"`
}

func CreateEvent(event Event) error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO events (ev_event_type, ev_description, ev_timestampt, ev_creator_id) VALUES (?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(event.EventType, event.Description, event.TimeStampt, event.CreatorId)
	if err != nil {
		return err
	}
	return err
}
