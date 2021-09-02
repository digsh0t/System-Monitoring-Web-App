package models

import (
	"fmt"
	"strconv"
	"strings"
)

type HistoryInfo struct {
	HistoryId        int    `json:"historyId"`
	HistoryTimeStamp string `json:"historyTime"`
	HistoryCommand   string `json:"historyCommand"`
}

func HistoryListAll(sshConnectionId int) ([]HistoryInfo, error) {
	var (
		historyList []HistoryInfo
		err         error
	)
	SshConnectionInfo, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return historyList, err
	}
	//command := "HISTTIMEFORMAT=\"%d/%m/%y %T \" && history"
	command := "history | less"
	result, err := RunCommandFromSSHConnection(*SshConnectionInfo, command)
	if err != nil {
		if !strings.Contains(err.Error(), "Process exited with status 2") {
			return historyList, err
		}
	}
	fmt.Println("result", result)
	lines := strings.Split(strings.TrimSpace(result), "\n")
	for _, line := range lines {

		var history HistoryInfo
		attributeHistory := strings.Split(line, " ")
		history.HistoryId, err = strconv.Atoi(attributeHistory[0])
		if err != nil {
			return historyList, err
		}
		date := strings.TrimSpace(attributeHistory[1])
		time := attributeHistory[2]
		history.HistoryTimeStamp = date + " " + time
		history.HistoryCommand = attributeHistory[3]
		historyList = append(historyList, history)
	}
	return historyList, err
}
