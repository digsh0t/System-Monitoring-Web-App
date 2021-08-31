package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func AddTemplate(w http.ResponseWriter, r *http.Request) {

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

	var template models.Template

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("invalid request body").Error())
		return
	}

	err = json.Unmarshal(reqBody, &template)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read template info").Error())
		return
	}

	err = template.AddTaskToDB()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to insert template to database").Error())
		return
	}
	utils.JSON(w, http.StatusCreated, nil)
}
