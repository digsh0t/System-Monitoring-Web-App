package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetReport(w http.ResponseWriter, r *http.Request, start time.Time) {

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

	report, err := models.GetReport(r, start)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, report)
	}

}

func GetDetailOSReport(w http.ResponseWriter, r *http.Request) {

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

	// Get Id parameter
	query := r.URL.Query()
	ostype := query.Get("ostype")

	report, err := models.GetDetailOSReport(ostype)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, report)
	}

}

func ExportReport(w http.ResponseWriter, r *http.Request) {

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

	// Retrieve Json Format
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to retrieve json format").Error())
		return
	}

	var modules models.ReportModules
	err = json.Unmarshal(reqBody, &modules)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to process json")
		return
	}

	// Get current date time
	datetime := utils.GetCurrentDateTime()
	filename := "./tmp/report-" + datetime + ".pdf"
	err = models.ExportReport(filename, modules)
	var eventStatus string
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		eventStatus = "failed"
	} else {
		// Send Email
		utils.EnvInit()
		sI := models.SmtpInfo{EmailSender: os.Getenv("EMAIL_SENDER"), EmailPassword: os.Getenv("EMAIL_PASS"), SMTPHost: os.Getenv("SMTP_HOST"), SMTPPort: os.Getenv("SMTP_PORT")}
		err = sI.SendReportMail(filename, modules.ReceiveEmail, modules.Cc, modules.Bcc, r)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}

		// Remove tmp pdf file
		err = os.Remove(filename)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.JSON(w, http.StatusOK, err)
		eventStatus = "successfulyy"

	}
	// Write Event Web
	description := "Export report " + eventStatus
	_, err = models.WriteWebEvent(r, "Template", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write template event").Error())
		return
	}

}
