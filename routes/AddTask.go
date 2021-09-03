package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/event"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func AddTask(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

	//Authorization
	isAuthorized, err := auth.CheckAuth(r, []string{"admin"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read request body").Error())
		return
	}

	var task models.Task
	json.Unmarshal(body, &task)

	task.UserId, err = auth.ExtractUserId(r)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read user id from token").Error())
		return
	}

	task.Status = "waiting"
	task.AddTaskToDB()
	// Write Event Web
	description := "Task Id \"" + strconv.Itoa(task.TaskId) + "\" waiting to run"
	_, err = event.WriteWebEvent(r, "Task", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write task event").Error())
		return
	}
	err = task.Run()
	// Write Event Web
	description = "Task Id \"" + strconv.Itoa(task.TaskId) + "\" finished with result: " + task.Status
	event.WriteWebEvent(r, "Task", description)

	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
}
