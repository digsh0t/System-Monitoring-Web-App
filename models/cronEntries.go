package models

import (
	"fmt"

	"github.com/robfig/cron/v3"
)

var RemoveEntryCh = make(chan cron.EntryID, 100)
var CurrentEntryCh = make(chan cron.EntryID, 100)
var C = cron.New()

func RemoveEntry(id int) {
	var cronId cron.EntryID = cron.EntryID(id)
	RemoveEntryCh <- cronId
}

func RemoveEntryChannel() {
	for {
		x, ok := <-RemoveEntryCh
		if ok {
			C.Remove(x)
		} else {
			fmt.Println("Channel closed!")
		}
	}
}

func CurrentEntryChannel() {
	for {
		x, ok := <-CurrentEntryCh
		if ok {
			fmt.Printf("Value %d was read to current cron channel.\n", x)
		} else {
			fmt.Println("Channel closed!")
		}
	}
}
