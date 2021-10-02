package models

import (
	"errors"
	"strconv"
	"strings"
)

type WindowsLogs struct {
	Index      int    `json:"index"`
	Time       string `json:"time"`
	EntryType  string `json:"entryType"`
	Source     string `json:"source"`
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

// Get Windows Event Logs API
func GetWindowsEventLogs(sshConnectionId int, logname string, startTime string, endTime string) ([]WindowsLogs, error) {
	var (
		windowsLogsList []WindowsLogs
		err             error
	)

	// Get Connection By SshConnectionId
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return windowsLogsList, errors.New("fail to get ssh connection")
	}

	// Prepare Query
	query := "PowerShell -Command Get-EventLog -LogName " + logname + " -Newest 100"
	if startTime != "" {
		query += " -After " + "'" + startTime + "'"
	}
	if endTime != "" {
		query += " -Before " + "'" + endTime + "'"
	}
	// Run remote command
	output, err := sshConnection.RunCommandFromSSHConnectionUseKeys(query)
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

// Get Detail Windows Event log
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

	// Trim Space and spilit lines
	output = strings.TrimSpace(output)
	lines := strings.Split(output, "\n")

	// Record current key
	var currentAttribute string
	for _, line := range lines {

		if line[19] == ':' {

			// Get Key and Value
			key := strings.TrimSpace(line[0:19])
			value := strings.TrimSpace(line[20:])

			// Switch each case to get value
			switch key {
			case "EventID":
				eventId, err := strconv.Atoi(value)
				if err != nil {
					return detailWindowsLog, err
				}
				detailWindowsLog.EventId = eventId
				currentAttribute = "EventID"
			case "MachineName":
				detailWindowsLog.MachineName = value
				currentAttribute = "MachineName"
			case "Data":
				detailWindowsLog.Data = value
				currentAttribute = "Data"
			case "Index":
				index, err := strconv.Atoi(value)
				if err != nil {
					return detailWindowsLog, err
				}
				detailWindowsLog.Index = index
				currentAttribute = "Index"
			case "CategoryNumber":
				categoryNumber, err := strconv.Atoi(value)
				if err != nil {
					return detailWindowsLog, err
				}
				detailWindowsLog.CategoryNumber = categoryNumber
				currentAttribute = "CategoryNumber"
			case "EntryType":
				detailWindowsLog.EntryType = value
				currentAttribute = "EntryType"
			case "Message":
				detailWindowsLog.Message = value
				currentAttribute = "Message"
			case "Source":
				detailWindowsLog.Source = value
				currentAttribute = "Source"
			case "ReplacementStrings":
				detailWindowsLog.ReplacementStrings = value
				currentAttribute = "ReplacementStrings"
			case "InstanceId":
				instanceId, err := strconv.Atoi(value)
				if err != nil {
					return detailWindowsLog, err
				}
				detailWindowsLog.InstanceId = instanceId
				currentAttribute = "InstanceId"
			case "TimeGenerated":
				detailWindowsLog.TimeGenerated = value
				currentAttribute = "TimeGenerated"
			case "UserName":
				detailWindowsLog.UserName = value
				currentAttribute = "UserName"
			}
		} else {

			// Special case needed to append
			value := line[21:] + "\n"
			switch currentAttribute {
			case "MachineName":
				detailWindowsLog.MachineName += value
			case "Data":
				detailWindowsLog.Data += value
			case "EntryType":
				detailWindowsLog.EntryType += value
			case "Message":
				detailWindowsLog.Message += value
			case "Source":
				detailWindowsLog.Source += value
			case "ReplacementStrings":
				detailWindowsLog.ReplacementStrings += value
			case "TimeGenerated":
				detailWindowsLog.TimeGenerated += value
			case "UserName":
				detailWindowsLog.UserName += value
			}
		}
	}

	// Return
	return detailWindowsLog, err
}
