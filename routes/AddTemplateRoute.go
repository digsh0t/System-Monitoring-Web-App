package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/albrow/forms"
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

	templateData, err := forms.Parse(r)
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to parse form-data").Error())
		return
	}

	template = models.Template{
		TemplateName: templateData.Get("template_name"),
		Description:  templateData.Get("template_description"),
		SshKeyId:     30,
		Alert:        templateData.GetBool("alert"),
	}

	// Process yaml file
	r.ParseMultipartForm(10 << 20)
	var buf bytes.Buffer
	file, handler, err := r.FormFile("yaml_file")
	yamlName := handler.Filename
	//Get yaml file content
	io.Copy(&buf, file)
	yamlContent := buf.String()
	buf.Reset()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to yaml file").Error())
		return
	}

	template.UserId, err = auth.ExtractUserId(r)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read user id from token").Error())
		return
	}

	lastIndex, err := template.AddTemplateToDB()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to insert template to database").Error())
		return
	}
	path := "./yamls/templates/" + strconv.Itoa(int(lastIndex)) + "-" + yamlName
	newFile, err := os.Create(path)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to create new file").Error())
		return
	}
	defer newFile.Close()

	_, err = newFile.WriteString(yamlContent)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write content to new file").Error())
		return
	}
	template.FilePath = path
	template.TemplateId = int(lastIndex)
	err = template.UpdateFilePath()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to update file path").Error())
		return
	}

	// Write Event Web
	description := "Task Id \"" + strconv.Itoa(template.TemplateId) + "\" created "
	_, err = models.WriteWebEvent(r, "Template", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write template event").Error())
		return
	}

	utils.JSON(w, http.StatusCreated, nil)
}
