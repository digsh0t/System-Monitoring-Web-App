package utils

import (
	"os/exec"
	"regexp"
	"strings"
)

func IsNodeReachable(clientname string) bool {
	rawOuput, _ := exec.Command("ansible", clientname, "-m", "ping").Output()
	ouput := string(rawOuput)
	for _, line := range strings.Split(strings.TrimRight(ouput, "\n"), "\n") {
		line = strings.TrimSpace(line)

		r, _ := regexp.Compile("SUCCESS")
		if r.MatchString(line) {
			return true
		}

	}
	return false
}
