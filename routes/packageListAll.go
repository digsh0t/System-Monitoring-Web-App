package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
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

	returnJson := simplejson.New()
	vars := mux.Vars(r)
	stringId := vars["id"]
	if stringId == "all" {
		PackageList, err = models.GetAllPackage()
		if err != nil {
			returnJson.Set("Status", false)
			returnJson.Set("Error", errors.New("Fail to retrieve from DB").Error())
			utils.JSON(w, http.StatusBadRequest, returnJson)
			return
		}
	} else {
		intId, err := strconv.Atoi(stringId)
		if err != nil {
			returnJson.Set("Status", false)
			returnJson.Set("Error", errors.New("invalid SSH Connection id").Error())
			utils.JSON(w, http.StatusBadRequest, returnJson)
			return
		}
		PackageList, err = models.GetAllPackageFromHostID(intId)
		if err != nil {
			returnJson.Set("Status", false)
			returnJson.Set("Error", errors.New("Fail to retrieve from DB").Error())
			utils.JSON(w, http.StatusBadRequest, returnJson)
			return
		}
	}

	// Return json

	if err != nil {
		utils.JSON(w, http.StatusBadRequest, PackageList)
		return
	}

	utils.JSON(w, http.StatusOK, PackageList)

}
