package models

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
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

func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix = true
		err      error
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

func logPipe(reader *bufio.Reader) {

	line, err := Readln(reader)
	for err == nil {
		fmt.Println("Debug: " + line)
		line, err = Readln(reader)
	}

	if err != nil && err.Error() != "EOF" {
		log.Println("fail to read task output")
	}
}

func logCmd(cmd *exec.Cmd) {
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	go logPipe(bufio.NewReader(stderr))
	go logPipe(bufio.NewReader(stdout))
}

func (template *Template) RunPlaybook() error {
	defer func() {
		finishedTime := time.Now()
		description := "Template Id " + strconv.Itoa(template.TemplateId) + " ( " + template.Description + " ) " + "finished at " + finishedTime.String()
		CreateEvent(Event{EventType: "Run Template", Description: description, TimeStampt: finishedTime, CreatorId: 1})
	}()

	cmd := exec.Command("ansible-playbook", template.FilePath)
	logCmd(cmd)
	cmd.Stdin = strings.NewReader("")
	return cmd.Run()
}

func (template *Template) AddTemplateToDB() error {
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
