package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func ClientExportReport(w http.ResponseWriter, r *http.Request) {

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
	err = models.ClientExportReport(filename, modules)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		/*file, err := ioutil.ReadFile(filename)
		if err != nil {
			utils.ERROR(w, http.StatusUnauthorized, err.Error())
			return
		}*/
		sI := models.SmtpInfo{EmailSender: "noti.lthmonitor@gmail.com", EmailPassword: "Lethihang123", SMTPHost: "smtp.gmail.com", SMTPPort: "587"}
		err = sI.SendReportMail(filename, []string{"longhkse140235@fpt.edu.vn"}, modules.Cc, modules.Bcc, r)
		if err != nil {
			utils.ERROR(w, http.StatusUnauthorized, err.Error())
			return
		}

		// Remove tmp pdf file
		/*err = os.Remove(filename)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}*/
		utils.JSON(w, http.StatusOK, err)

	}

}
