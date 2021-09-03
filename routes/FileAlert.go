package routes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/wintltr/login-api/alerts"
	"github.com/wintltr/login-api/utils"
)

func WatchFile(w http.ResponseWriter, r *http.Request) {
	type file struct {
		Filepath string `json:"filepath"`
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	}

	var f file
	json.Unmarshal(body, &f)

	go alerts.WatchFile(f.Filepath)

}
