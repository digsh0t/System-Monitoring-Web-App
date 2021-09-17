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

func PackageListAll(w http.ResponseWriter, r *http.Request) {

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

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to retrieve Json format").Error())
		return
	}
	// Parse Json
	var packageJson models.PackageJson
	json.Unmarshal(reqBody, &packageJson)

	// Called List All Package
	packageList, err := models.ListAllPackge(packageJson.Host)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to get installed package").Error())
		return
	}

	// Return json
	utils.JSON(w, http.StatusOK, packageList)

}
