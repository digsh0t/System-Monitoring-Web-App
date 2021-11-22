package routes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func AddNewWatcherRoute(w http.ResponseWriter, r *http.Request) {
	var watch models.WatchInfo
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	err = json.Unmarshal(body, &watch)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	err = watch.AddNewWatcher()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusCreated, watch)
}

func GetAllWatcherRoute(w http.ResponseWriter, r *http.Request) {
	watch, err := models.GetAllWatch()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, watch)
}

func RemoveFromWatchRoute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sshConnectionId, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	err = models.RemoveWatcher(sshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, nil)
}
