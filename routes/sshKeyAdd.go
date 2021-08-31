package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/event"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func AddSSHKey(w http.ResponseWriter, r *http.Request) {

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

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	var buf bytes.Buffer
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("privateKey")
	//Get Key file name
	keyName := handler.Filename
	//Get Key file content
	io.Copy(&buf, file)
	privateKey := buf.String()
	buf.Reset()
	returnJson := simplejson.New()
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("error while processing key file").Error())
		utils.JSON(w, http.StatusServiceUnavailable, returnJson)
		return
	}
	defer file.Close()

	creatorId, err := auth.ExtractUserId(r)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", err.Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	var sshKey models.SSHKey
	sshKey.KeyName = keyName
	sshKey.PrivateKey = models.AESEncryptKey(privateKey)
	sshKey.CreatorId = creatorId

	status, err := sshKey.InsertSSHKeyToDB()
	returnJson.Set("Status", status)
	statusCode := http.StatusBadRequest
	if err != nil {
		returnJson.Set("Error", err.Error())
	} else {
		returnJson.Set("Error", err)
		statusCode = http.StatusCreated
	}

	// Write Event Web
	description := "Add sshKey " + sshKey.KeyName + " to DB"
	_, err = event.WriteWebEvent(r, "SSHKey", description)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to write web event")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	utils.JSON(w, statusCode, returnJson)
}
