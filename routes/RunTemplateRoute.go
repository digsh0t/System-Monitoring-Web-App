package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func RunTemplate(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	templateId, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("invalid template id").Error())
		return
	}

	template, err := models.GetTemplateFromId(templateId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read template from database").Error())
		return
	}

	err = template.RunPlaybook()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("invalid request body").Error())
		return
	}

	utils.JSON(w, http.StatusCreated, nil)
}
