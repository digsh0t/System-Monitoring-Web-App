package models

import (
	"bufio"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/wintltr/login-api/database"
)

type Template struct {
	TemplateId   int    `json:"template_id"`
	TemplateName string `json:"template_name"`
	Description  string `json:"template_description"`
	SshKeyId     int    `json:"ssh_key_id"`
	FilePath     string `json:"filepath"`
	Arguments    string `json:"arguments"`
	Alert        bool   `json:"alert"`
	UserId       int    `json:"user_id"`
	Username     string `json:"username"`
}

func (template *Template) AddTemplateToDB() (int64, error) {
	var lastIndex int64
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO templates (template_name, template_description, ssh_key_id, filepath, arguments, alert, user_id) VALUES (?,?,?,?,?,?,?)")
	if err != nil {
		return lastIndex, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(template.TemplateName, template.Description, template.SshKeyId, template.FilePath, template.Arguments, template.Alert, template.UserId)
	if err != nil {
		return lastIndex, err
	}
	lastIndex, err = res.LastInsertId()
	if err != nil {
		return lastIndex, err
	}
	return lastIndex, err
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

func GetAllTemplate() ([]Template, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT T.template_id, T.template_name, T.template_description, T.ssh_key_id, T.filepath, T.arguments, T.alert, T.user_id, WA.wa_users_username FROM templates AS T LEFT JOIN wa_users AS WA ON T.user_id = WA.wa_users_id`
	selDB, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var template Template
	var templateList []Template
	for selDB.Next() {
		err = selDB.Scan(&template.TemplateId, &template.TemplateName, &template.Description, &template.SshKeyId, &template.FilePath, &template.Arguments, &template.Alert, &template.UserId, &template.Username)
		if err != nil {
			return nil, err
		}

		templateList = append(templateList, template)
	}
	return templateList, err
}

func DeleteTemplateFromId(templateId int) error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM templates WHERE template_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(templateId)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return errors.New("no template with id: " + strconv.Itoa(templateId) + " exists")
	}
	return err
}

// Update Os Type to DB
func (template *Template) UpdateFilePath() error {
	db := database.ConnectDB()
	defer db.Close()

	query := "UPDATE templates SET filepath = ? WHERE template_id = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(template.FilePath, template.TemplateId)
	if err != nil {
		return err
	}
	return err
}

func (template *Template) GetTemplateArgument() ([]string, error) {
	var (
		arguments []string
		err       error
	)
	file, err := os.Open(template.FilePath)
	if err != nil {
		return arguments, errors.New("fail to read file content")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		regex, _ := regexp.Compile(`\{\{.*?\}\}`)
		argument := regex.FindString(scanner.Text())
		if argument != "" {
			replacer := strings.NewReplacer("{", "", "}", "", " ", "")
			argument = replacer.Replace(argument)
			if !strings.Contains(argument, "ansible") && !strings.Contains(argument, "item") {
				arguments = append(arguments, argument)
			}
		}

	}
	return arguments, err

}
