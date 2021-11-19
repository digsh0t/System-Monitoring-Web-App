package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bitly/go-simplejson"
)

func ReturnInsertJSON(w http.ResponseWriter, status bool, err error) {
	returnJson := simplejson.New()
	returnJson.Set("Status", status)
	statusCode := http.StatusBadRequest
	if err != nil {
		returnJson.Set("Error", err.Error())
	} else {
		returnJson.Set("Error", err)
		statusCode = http.StatusCreated
	}
	JSON(w, statusCode, returnJson)
}

func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

func ERROR(w http.ResponseWriter, statusCode int, err string) {
	if err != "" {
		JSON(w, statusCode, struct {
			Error string `json:"error"`
		}{
			Error: err,
		})
		return
	}
	JSON(w, http.StatusBadRequest, nil)
}
