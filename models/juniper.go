package models

import (
	"errors"
	"regexp"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/wintltr/login-api/utils"
)

// Get Log Cisco
func ParseLogsJuniper(output string) ([]NetworkLogs, error) {
	var networkLogsList []NetworkLogs

	// Get substring from ansible output
	data := utils.ExtractSubString(output, " => ", "PLAY RECAP")

	// Parse Json format
	jsonParsed, err := gabs.ParseJSON([]byte(data))
	if err != nil {
		return networkLogsList, errors.New("fail to parse json output")
	}

	// Get List Arrays
	tmpList, err := jsonParsed.Search("msg").Children()
	if err != nil {
		return networkLogsList, errors.New("fail to parse json output")
	}

	// Get Specific Array
	lines, err := tmpList[0].Children()
	if err != nil {
		return networkLogsList, errors.New("fail to parse json output")
	}

	// Line: "Oct 20 10:06:22   i386_junos[7105]: fpc0 connect failed"
	for _, line := range lines {
		var networkLogs NetworkLogs

		// Special case
		if strings.Contains(line.String(), "more 100%") || strings.Contains(line.String(), "                               ") {
			continue
		}

		// Special case
		if strings.Contains(line.String(), "last message repeated") {
			networkLogs.TimeStamp = strings.Trim(line.String()[:16], "\"")
			networkLogs.Description = strings.Trim(line.String()[19:], "\"")
			networkLogsList = append(networkLogsList, networkLogs)
			continue
		}

		// Check if existing log, case no returns empty list
		if line.String() == "\"\"" {
			return networkLogsList, nil
		}

		// Get Time
		networkLogs.TimeStamp = strings.Trim(line.String()[:16], "\"")

		// Get Service and Description
		tmpLine := line.String()[19:]
		re := regexp.MustCompile(":")
		spilit := re.Split(tmpLine, 2)

		networkLogs.Service = spilit[0]

		networkLogs.Description = strings.TrimSpace(spilit[1])

		// Append to list
		networkLogsList = append(networkLogsList, networkLogs)

	}

	return networkLogsList, err

}
