package models

import (
	"errors"
	"strconv"

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
	UserId       int    `json:"user_id"`
}

func (template *Template) AddTemplateToDB() error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO templates (template_name, template_description, ssh_key_id, filepath, arguments, alert, user_id) VALUES (?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(template.TemplateName, template.Description, template.SshKeyId, template.FilePath, template.Arguments, template.Alert, template.UserId)
	return err
}

func GetTemplateFromId(temlateId int) (Template, error) {
	db := database.ConnectDB()
	defer db.Close()

	var template Template
	row := db.QueryRow("SELECT template_id, template_name, template_description, ssh_key_id, filepath, arguments, alert FROM templates WHERE template_id = ?", temlateId)
	if row == nil {
		return Template{}, errors.New("no template with id " + strconv.Itoa(temlateId) + " exists")
	}
	err := row.Scan(&template.TemplateId, &template.TemplateName, &template.Description, &template.SshKeyId, &template.FilePath, &template.Arguments, &template.Alert)
	return template, err
}
