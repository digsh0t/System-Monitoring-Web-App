package routes

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/wintltr/login-api/utils"
)

func AddNewSSHConnection(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Println(string(body))
	utils.JSON(w, http.StatusOK, body)
}
