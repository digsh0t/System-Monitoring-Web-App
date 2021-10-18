package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

// Get Cisco Traffic
func GetRouterInterfaces(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.Atoi(query.Get("id"))
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to convert id").Error())
		return
	}

	interfaceSNMP, err := models.GetRouterInterfaces(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, interfaceSNMP)
	}

}

// Get Cisco Traffic
func GetRouterSystem(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.Atoi(query.Get("id"))
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to convert id").Error())
		return
	}

	systemSNMP, err := models.GetRouterSystem(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, systemSNMP)
	}

}

// Get Router IP Info
func GetRouterIPAddr(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.Atoi(query.Get("id"))
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to convert id").Error())
		return
	}

	ipSNMP, err := models.GetRouterIPAddr(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, ipSNMP)
	}

}

// Get Router IP Info
func GetRouterIPNetToMedia(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.Atoi(query.Get("id"))
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to convert id").Error())
		return
	}

	ipSNMP, err := models.GetRouterIPNetToMedia(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, ipSNMP)
	}

}

// Get Router IP Info
func GetRouterIPRoute(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.Atoi(query.Get("id"))
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to convert id").Error())
		return
	}

	routeSNMP, err := models.GetRouterIPRoute(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, routeSNMP)
	}

}
