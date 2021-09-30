package models

import (
	"errors"
	"strconv"
	"strings"
)

type WindowsLogs struct {
	Index      int    `json:"Index"`
	Time       string `json:"Time"`
	EntryType  string `json:"EntryType"`
	Source     string `json:"Source"`
	InstanceId int    `json:"instanceId"`
	Message    string `json:"message"`
}

type DetailWindowsLog struct {
	EventId            int    `json:"eventId"`
	MachineName        string `json:"machineName"`
	Data               string `json:"data"`
	Index              int    `json:"index"`
	CategoryNumber     int    `json:"categoryNumber"`
	EntryType          string `json:"entryType"`
	Message            string `json:"message"`
	Source             string `json:"source"`
	ReplacementStrings string `json:"replacementStrings"`
	InstanceId         int    `json:"instanceId"`
	TimeGenerated      string `json:"timeGenerated"`
	UserName           string `json:"userName"`
}

// Get Windows Event logs
func GetWindowsEventLogs(sshConnectionId int, logname string) ([]WindowsLogs, error) {
	var (
		windowsLogsList []WindowsLogs
		err             error
	)

	// Get Connection
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return windowsLogsList, errors.New("fail to get ssh connection")
	}

	// Run remote command
	output, err := sshConnection.RunCommandFromSSHConnectionUseKeys("PowerShell -Command Get-EventLog -LogName " + logname + " -Newest 100")
	if err != nil {
		return windowsLogsList, errors.New("fail to get windows event log")
	}

	// No record
	if len(output) == 0 {
		return windowsLogsList, nil
	}

	output = strings.TrimSpace(output)

	// Spilit line
	lines := strings.Split(output, "\n")
	for index, line := range lines {
		var windowsLog WindowsLogs
		if index > 2 {

			// Get Log Index
			indexRaw := line[:8]
			windowsLog.Index, err = strconv.Atoi(strings.TrimSpace(indexRaw))
			if err != nil {
				return windowsLogsList, err
			}

			// Get Log Time
			windowsLog.Time = line[9:21]

			// Get Log EntryType
			windowsLog.EntryType = strings.TrimSpace(line[23:34])

			// Get Log Source
			windowsLog.Source = strings.TrimSpace(line[35:55])

			// Get Log InstanceId
			InstanceIdRaw := line[56:68]
			windowsLog.InstanceId, err = strconv.Atoi(strings.TrimSpace(InstanceIdRaw))
			if err != nil {
				return windowsLogsList, err
			}

			// Get Log Message
			messageRaw := line[69:]
			windowsLog.Message = strings.TrimRight(messageRaw, "\r ")

			// Append to List
			windowsLogsList = append(windowsLogsList, windowsLog)

		}
	}
	return windowsLogsList, err
}

// Get Windows Event logs
func GetDetailWindowsEventLog(sshConnectionId int, logname string, index string) (DetailWindowsLog, error) {
	var (
		detailWindowsLog DetailWindowsLog
		err              error
	)

	// Get Connection
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return detailWindowsLog, errors.New("fail to get ssh connection")
	}

	// Run remote command
	output, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`PowerShell -Command "Get-EventLog -LogName ` + logname + ` -Index ` + index + ` | Select-Object -Property *"`)
	if err != nil {
		return detailWindowsLog, errors.New("fail to get detail windows event log")
	}

	// No record
	if len(output) == 0 {
		return detailWindowsLog, nil
	}

	output = strings.TrimSpace(output)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		attributes := strings.Split(line, ":")
		value := strings.TrimSpace(attributes[1])
		switch strings.TrimSpace(attributes[0]) {
		case "EventID":
			eventId, err := strconv.Atoi(value)
			if err != nil {
				return detailWindowsLog, err
			}
			detailWindowsLog.EventId = eventId
		case "MachineName":
			detailWindowsLog.MachineName = value
		case "Data":
			detailWindowsLog.Data = value
		case "Index":
			index, err := strconv.Atoi(value)
			if err != nil {
				return detailWindowsLog, err
			}
			detailWindowsLog.Index = index
		case "CategoryNumber":
			categoryNumber, err := strconv.Atoi(value)
			if err != nil {
				return detailWindowsLog, err
			}
			detailWindowsLog.CategoryNumber = categoryNumber
		case "EntryType":
			detailWindowsLog.EntryType = value
		case "Message":
			detailWindowsLog.Message = value
		case "Source":
			detailWindowsLog.Source = value
		case "ReplacementStrings":
			detailWindowsLog.ReplacementStrings = value
		case "InstanceId":
			instanceId, err := strconv.Atoi(value)
			if err != nil {
				return detailWindowsLog, err
			}
			detailWindowsLog.InstanceId = instanceId
		case "TimeGenerated":
			detailWindowsLog.TimeGenerated = value
		case "UserName":
			detailWindowsLog.UserName = value
		}
	}

	return detailWindowsLog, err
}
