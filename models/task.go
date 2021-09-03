package models

import (
	"bufio"
	"errors"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/wintltr/login-api/database"
)

type Task struct {
	TaskId        int       `json:"task_id"`
	TemplateId    int       `json:"template_id"`
	OverridedArgs string    `json:"overrided_args"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        string    `json:"status"`
	Alert         bool      `json:"alert"`
	UserId        int       `json:"user_id"`
}

type TaskResult struct {
	TaskId    int       `json:"task_id"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
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

func (task *Task) logPipe(reader *bufio.Reader) {

	line, err := Readln(reader)
	for err == nil {
		task.Log(line)
		line, err = Readln(reader)
	}

	if err != nil && err.Error() != "EOF" {
		log.Println("fail to read task output")
	}
}

func (task *Task) logCmd(cmd *exec.Cmd) {
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	go task.logPipe(bufio.NewReader(stderr))
	go task.logPipe(bufio.NewReader(stdout))
}

func (task *Task) Log(message string) error {
	//If message emtpy, won't log to DB
	if message == "" {
		return nil
	}
	err := LogTaskResult(TaskResult{
		TaskId:    task.TaskId,
		Timestamp: time.Now(),
		Message:   message,
	})
	return err
}

func LogTaskResult(taskResult TaskResult) error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO task_results (task_id, timestamp, message) VALUES (?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(taskResult.TaskId, taskResult.Timestamp, taskResult.Message)
	return err
}

func (task *Task) UpdateTask() error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("UPDATE tasks SET status = ?, start_time = ?, end_time = ? WHERE task_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(task.Status, task.StartTime, task.EndTime, task.TaskId)
	return err
}

func (task *Task) GetLatestTaskId() error {
	db := database.ConnectDB()
	defer db.Close()

	row := db.QueryRow("SELECT task_id FROM tasks ORDER BY task_id DESC LIMIT 1;")
	err := row.Scan(&task.TaskId)
	if row == nil {
		return errors.New("you have not created any task yet")
	}

	return err
}

func (task *Task) AddTaskToDB() error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO tasks (template_id, overrided_args, status, user_id) VALUES (?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(task.TemplateId, task.OverridedArgs, task.Status, task.UserId)
	if err != nil {
		return err
	}

	err = task.GetLatestTaskId()
	return err
}

func (task *Task) RunPlaybook() error {

	template, err := GetTemplateFromId(task.TemplateId)
	if err != nil {
		return errors.New("fail to get template of task")
	}

	//If task provide no arguments, use template's argument
	if task.OverridedArgs == "" {
		task.OverridedArgs = template.Arguments
	}

	args, _ := task.PrepareArgs()
	args = append(args, template.FilePath)

	cmd := exec.Command("ansible-playbook", args...)
	task.logCmd(cmd)
	cmd.Stdin = strings.NewReader("")
	return cmd.Run()
}

func (task *Task) PrepareArgs() ([]string, error) {
	var args []string
	if task.OverridedArgs != "" {
		args = append(args, "--extra-vars="+task.OverridedArgs)
	}
	return args, nil
}

func (task *Task) Run() error {
	task.Status = "running"
	task.Log("Task Id: " + strconv.Itoa(task.TaskId) + " run using Template Id: " + strconv.Itoa(task.TemplateId))
	task.StartTime = time.Now()
	err := task.RunPlaybook()
	task.EndTime = time.Now()

	if err != nil && !strings.Contains(err.Error(), "fail to read task output") {
		task.Log("Task Id: " + strconv.Itoa(task.TaskId) + " failed")
		task.Status = "failed"
		if task.Alert {
			SendTelegramMessage(task.EndTime.String() + ": Task Id " + strconv.Itoa(task.TaskId) + " is finished with ERROR")
		}
	} else {
		task.Log("Task Id: " + strconv.Itoa(task.TaskId) + " has ran succesfully")
		task.Status = "success"
	}

	task.UpdateTask()
	return err
}

func GetTaskLog(taskId int) ([]TaskResult, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT task_id, timestamp, message FROM task_results WHERE task_id = ?`
	selDB, err := db.Query(query, taskId)
	if err != nil {
		return nil, err
	}

	var taskResult TaskResult
	var manyTaskResults []TaskResult
	for selDB.Next() {
		err = selDB.Scan(&taskResult.TaskId, &taskResult.Timestamp, &taskResult.Message)
		if err != nil {
			return nil, errors.New("fail to get task results from Database")
		}

		manyTaskResults = append(manyTaskResults, taskResult)
	}
	return manyTaskResults, err
}
