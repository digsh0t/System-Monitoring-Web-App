package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func PackageListAll(w http.ResponseWriter, r *http.Request) {
	var (
		PackageList []models.PackageInstalledInfo
		err         error
	)

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

	vars := mux.Vars(r)
	stringId := vars["id"]
	if stringId == "all" {
		PackageList, err = models.GetAllPackage()
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to retrieve from db").Error())
			return
		}
	} else {
		intId, err := strconv.Atoi(stringId)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, errors.New("Invalid SshConnectionId").Error())
			return
		}
		PackageList, err = models.GetAllPackageFromHostID(intId)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to retrieve from db").Error())
			return
		}
	}

	// Return json
	utils.JSON(w, http.StatusOK, PackageList)

}
