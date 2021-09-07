package models

import (
	"bufio"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/wintltr/login-api/database"
)

type Task struct {
	TaskId        int       `json:"task_id"`
	TemplateId    int       `json:"template_id"`
	OverridedArgs string    `json:"overrided_args"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	CronTime      string    `json:"cron_time"`
	Status        string    `json:"status"`
	Alert         bool
	UserId        int `json:"user_id"`
	CronId        cron.EntryID
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
	queryString := "UPDATE tasks SET status = ?, start_time = ?, end_time = ?, cron_id=0 WHERE task_id = ?"
	if task.StartTime.IsZero() && task.EndTime.IsZero() {
		queryString = "UPDATE tasks SET status = ?, cron_id=0 WHERE task_id = ?"
	} else if task.EndTime.IsZero() {
		queryString = "UPDATE tasks SET status = ?, start_time = ?, cron_id=0 WHERE task_id = ?"
	}
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare(queryString)
	if err != nil {
		return err
	}
	defer stmt.Close()
	if task.StartTime.IsZero() && task.EndTime.IsZero() {
		_, err = stmt.Exec(task.Status, task.TaskId)
	} else if task.EndTime.IsZero() {
		_, err = stmt.Exec(task.Status, task.StartTime, task.TaskId)
	} else {
		_, err = stmt.Exec(task.Status, task.StartTime, task.EndTime, task.TaskId)
	}
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

func (task *Task) AddCronTaskToDB() error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO tasks (template_id, overrided_args, start_time, status, cron_id, user_id) VALUES (?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(task.TemplateId, task.OverridedArgs, task.StartTime, task.Status, task.CronId, task.UserId)
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
		task.Status = "failed"
	} else {
		task.Status = "success"
	}

	task.UpdateStatus()
	return err
}

func (task *Task) UpdateStatus() error {
	task.Log("Task Id:" + strconv.Itoa(task.TaskId) + " run " + task.Status)
	if task.Alert {
		SendTelegramMessage(task.EndTime.String() + ": Task Id " + strconv.Itoa(task.TaskId) + " is finished with result: " + task.Status)
	}
	return task.UpdateTask()
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

func GetAllTasks(templateId int) ([]Task, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT task_id, template_id, overrided_args, start_time, end_time, status, user_id FROM tasks WHERE template_id = ?`
	selDB, err := db.Query(query, templateId)
	if !selDB.Next() {
		return nil, errors.New("template id not exists")
	}
	if err != nil {
		return nil, err
	}
	var startTime, endTime sql.NullTime
	var task Task
	var taskList []Task
	for selDB.Next() {
		err = selDB.Scan(&task.TaskId, &task.TemplateId, &task.OverridedArgs, &startTime, &endTime, &task.Status, &task.UserId)
		if err != nil {
			return nil, errors.New("fail to get tasks from database with template id: " + strconv.Itoa(templateId))
			//return nil, err
		}
		if startTime.Valid && endTime.Valid {
			task.StartTime = startTime.Time
			task.EndTime = endTime.Time
		}
		taskList = append(taskList, task)
		task.StartTime = time.Time{}
		task.EndTime = time.Time{}
	}
	return taskList, err
}

func (task *Task) Prepare(r *http.Request, startTime time.Time) error {
	task.Status = "waiting"
	var err error
	if task.CronTime != "" {
		task.StartTime = startTime
		err = task.AddCronTaskToDB()
		if err != nil {
			task.Status = "failed"
			task.UpdateStatus()
			return errors.New("Fail to write task event")
		}
		return err
	} else {
		task.AddTaskToDB()
		// Write Event Web
		description := "Task Id \"" + strconv.Itoa(task.TaskId) + "\" waiting to run"
		_, err := WriteWebEvent(r, "Task", description)
		if err != nil {
			task.Status = "failed"
			task.UpdateStatus()
			return errors.New("Fail to write task event")
		}
		return err
	}
}

func (task *Task) RunTask(r *http.Request) error {
	err := task.Prepare(r, time.Now())
	if err != nil {
		return err
	}
	err = task.Run()
	// Write Event Web
	description := "Task Id \"" + strconv.Itoa(task.TaskId) + "\" finished with result: " + task.Status
	WriteWebEvent(r, "Task", description)

	return err
}

func (task *Task) CronRunTask(r *http.Request) error {
	// id, _ := C.AddFunc(task.CronTime, func() { task.RunTask(r) })
	var err error
	var nextRun time.Time
	var isNewRun bool = true
	id, _ := C.AddFunc(task.CronTime, func() {
		err = task.Run()
		// Write Event Web
		description := "Task Id \"" + strconv.Itoa(task.TaskId) + "\" finished with result: " + task.Status
		WriteWebEvent(r, "Task", description)
		isNewRun = true
	})
	CurrentEntryCh <- id
	task.CronId = id
	C.Start()
	defer C.Stop()
	if err != nil {
		return err
	}
	for {
		time.Sleep(time.Second)
		if !C.Entry(id).Valid() {

			task.Status = "halted"
			description := "Task Id \"" + strconv.Itoa(task.TaskId) + "\" finished with result: " + task.Status
			WriteWebEvent(r, "Task", description)
			return task.UpdateStatus()

		} else if nextRun != C.Entry(id).Next && isNewRun { //If next run has changed (cron run is repeating), add new event log and update nextRun
			err = task.Prepare(r, C.Entry(id).Next)
			description := "Task id \" " + strconv.Itoa(task.TaskId) + "\" schedule to run at: " + C.Entry(id).Next.String()
			WriteWebEvent(r, "Task", description)
			nextRun = C.Entry(id).Next

			isNewRun = false
		}
	}
}
