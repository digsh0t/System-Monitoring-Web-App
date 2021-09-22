package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetAllSSHKey(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

	//Authorization
	isAuthorized, err := auth.CheckAuth(r, []string{"admin", "user"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	sshKeyList, err := models.GetAllSSHKeyFromDB()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	}

	utils.JSON(w, http.StatusOK, sshKeyList)

}

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
	var eventStatus string
	if err != nil {
		returnJson.Set("Error", err.Error())
		eventStatus = "failed"
	} else {
		returnJson.Set("Error", err)
		statusCode = http.StatusCreated
		eventStatus = "successfully"
	}

	utils.JSON(w, statusCode, returnJson)

	// Write Event Web
	description := "Add sshKey \"" + sshKey.KeyName + "\" to DB " + eventStatus
	_, err = models.WriteWebEvent(r, "SSHKey", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write event").Error())
		return
	}
}

func SSHKeyDeleteRoute(w http.ResponseWriter, r *http.Request) {

	// w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
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

	returnJson := simplejson.New()
	vars := mux.Vars(r)
	sshKeyId, err := strconv.Atoi(vars["id"])
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("invalid SSH key id").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	// Get SSHKey name
	sshKey, err := models.GetSSHKeyFromId(sshKeyId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Failed to get sshkey").Error())
		return
	}

	_, err = models.SSHKeyDelete(sshKeyId)
	var eventStatus string
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("please delete ssh connections associated with this SSH key first").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		eventStatus = "failed"
	} else {
		returnJson.Set("Status", true)
		returnJson.Set("Error", nil)
		utils.JSON(w, http.StatusOK, returnJson)
		eventStatus = "successfully"
	}

	// Write Event Web
	description := "Delete sshKey \"" + sshKey.KeyName + "\" from DB " + eventStatus
	_, err = models.WriteWebEvent(r, "SSHKey", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write event").Error())
		return
	}

}
