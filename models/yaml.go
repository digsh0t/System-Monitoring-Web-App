package models

import (
	"bytes"
	"fmt"
	"os/exec"
)

func RunAnsiblePlaybookWithjson(extraVars string, filepath string) error {

	var args []string
	if extraVars != "" {
		args = append(args, "--extra-vars", extraVars, filepath)
	} else {
		args = append(args, filepath)
	}

	cmd := exec.Command("ansible-playbook", args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}
	return err
}
