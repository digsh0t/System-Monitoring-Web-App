package routes

import (
	"net/http"
	"os/exec"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/utils"
)

func LoadFile(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	yamlName := vars["name"]
	command := exec.Command("ansible-playbook", "./yamls/"+yamlName)
	err := command.Run()
	if err != nil {
	}
	returnJson := simplejson.New()
	returnJson.Set("Status", true)
	returnJson.Set("Error", nil)
	utils.JSON(w, http.StatusOK, returnJson)

}
