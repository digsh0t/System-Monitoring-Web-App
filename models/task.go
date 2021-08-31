package models

import (
	"os/exec"
	"strconv"
	"time"

	"github.com/wintltr/login-api/database"
)

type Template struct {
	TemplateId   int    `json:"Template_id"`
	TemplateName string `json:"Template_name"`
	Description  string `json:"description"`
	SshKeyId     int    `json:"ssh_key_id"`
	FilePath     string `json:"filepath"`
	Arguments    string `json:"arguments"`
	Alert        bool   `json:"alert"`
}

func (template *Template) RunPlaybook() error {
	defer func() {
		finishedTime := time.Now()
		description := "Template Id " + strconv.Itoa(template.TemplateId) + " ( " + template.Description + " ) " + "finished at " + finishedTime.String()
		CreateEvent(Event{EventType: "Run Template", Description: description, TimeStampt: finishedTime, CreatorId: 1})
	}()

	cmd := exec.Command("ansible-playbook", template.FilePath)
	return cmd.Run()
}

func (template *Template) AddTaskToDB() error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO templates (template_name, template_description, ssh_key_id, filepath, arguments, alert) VALUES (?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(template.TemplateName, template.Description, template.SshKeyId, template.FilePath, template.Arguments, template.Alert)
	return err
}
