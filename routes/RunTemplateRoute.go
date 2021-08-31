package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func RunTemplate(w http.ResponseWriter, r *http.Request) {
	var template models.Template

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("invalid request body").Error())
		return
	}

	json.Unmarshal(reqBody, &template)
	err = template.RunPlaybook()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("invalid request body").Error())
		return
	}

	utils.JSON(w, http.StatusCreated, nil)
}
