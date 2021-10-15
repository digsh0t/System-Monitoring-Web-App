package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bitly/go-simplejson"
	"github.com/dgryski/dgoogauth"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GenerateQR(w http.ResponseWriter, r *http.Request) {

	//w.Write([]byte(fmt.Sprintf("Generating QR code\n")))
	userdata, err := auth.ExtractTokenMetadata(r)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	authLink, secret, err := models.GenerateQR(userdata.Username)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := models.GetUserByIdFromDB(userdata.Userid)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	err = user.UpdateSecret(secret)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	returnJson := simplejson.New()
	returnJson.Set("url", authLink)
	returnJson.Set("secret", secret)
	utils.JSON(w, http.StatusOK, returnJson)
}

func VerifyQR(w http.ResponseWriter, r *http.Request) {
	var twofa string

	userdata, err := auth.ExtractTokenMetadata(r)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	if userdata.Twofa == "not configured" {
		utils.ERROR(w, http.StatusBadRequest, errors.New("You have not configured 2FA!").Error())
		return
	}

	user, err := models.GetUserByIdFromDB(userdata.Userid)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	encrypted, err := user.GetSecret()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	secret, err := models.AESDecryptKey(encrypted)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	type tmpObj struct {
		Totp string `json:"totp"`
	}
	var otp tmpObj
	json.Unmarshal(body, &otp)

	// setup the one-time-password configuration.
	otpConfig := &dgoogauth.OTPConfig{
		Secret:      strings.TrimSpace(secret),
		WindowSize:  3,
		HotpCounter: 0,
	}

	trimmedToken := strings.TrimSpace(otp.Totp)

	// Validate token
	ok, err := otpConfig.Authenticate(trimmedToken)

	// if the token is invalid or expired
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	if !ok {
		user.Role = "not authorized"
		twofa = "not authorized"
	} else {
		twofa = "authorized"
	}

	token, err := auth.CreateToken(user.UserId, user.Username, user.Role, twofa)
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, "Fail to create token while login")
		return
	}

	returnJson := simplejson.New()
	returnJson.Set("ok", ok)
	returnJson.Set("authorization", token)
	utils.JSON(w, http.StatusOK, returnJson)
}

func VerifyQRSettingsRoute(w http.ResponseWriter, r *http.Request) {
	type verifyInfo struct {
		Password string `json:"password"`
		Secret   string `json:"secret"`
	}

	//w.Write([]byte(fmt.Sprintf("Generating QR code\n")))
	userdata, err := auth.ExtractTokenMetadata(r)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	var vI verifyInfo
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	err = json.Unmarshal(body, &vI)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := models.GetUserByIdFromDB(userdata.Userid)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	hashedPwd := models.HashPassword(vI.Password)
	secret, err := models.AESDecryptKey(user.Secret)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	returnJson := simplejson.New()
	if user.Password == hashedPwd && vI.Secret == secret {
		user.Update2FAStatus(true)
		returnJson.Set("success", true)
	} else {
		returnJson.Set("success", false)
	}
	utils.JSON(w, http.StatusOK, returnJson)
}

func TurnOff2FARoute(w http.ResponseWriter, r *http.Request) {

	type passwordInfo struct {
		Password string `json:"password"`
	}
	var pI passwordInfo
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	err = json.Unmarshal(body, &pI)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	//w.Write([]byte(fmt.Sprintf("Generating QR code\n")))
	userdata, err := auth.ExtractTokenMetadata(r)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := models.GetUserByIdFromDB(userdata.Userid)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	if user.Password == models.HashPassword(pI.Password) {
		err = user.UpdateSecret("")
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}
		err = user.Update2FAStatus(false)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}
	} else {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Wrong password!").Error())
		return
	}
}
