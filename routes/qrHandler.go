package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GenerateQR(w http.ResponseWriter, r *http.Request) {

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

	// Write Event Web
	description := "qr was generated"
	_, err = models.WriteWebEvent(r, "Windows", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}
}

func VerifyQR(w http.ResponseWriter, r *http.Request) {

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
	var twofa string

	userdata, err := auth.ExtractTokenMetadata(r)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	if userdata.Twofa == "not configured" {
		utils.ERROR(w, http.StatusBadRequest, errors.New("you have not configured 2FA!").Error())
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
	ok, err := models.CheckTOTP(secret, otp.Totp)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	if !ok {
		user.Role = "not authorized"
		twofa = "not authorized"
		utils.ERROR(w, http.StatusBadRequest, errors.New("Not authorized").Error())
		return
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
	returnJson.Set("user_id", user.UserId)
	utils.JSON(w, http.StatusOK, returnJson)
}

func VerifyQRSettingsRoute(w http.ResponseWriter, r *http.Request) {

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
	ok, err := models.CheckTOTP(secret, vI.Secret)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	returnJson := simplejson.New()
	if user.Password == hashedPwd && ok {
		user.Update2FAStatus(true)
		returnJson.Set("success", true)
	} else {
		returnJson.Set("success", false)
	}
	utils.JSON(w, http.StatusOK, returnJson)
}

func TurnOff2FARoute(w http.ResponseWriter, r *http.Request) {
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
		utils.ERROR(w, http.StatusBadRequest, errors.New("wrong password").Error())
		return
	}

	// Write Event Web
	description := "qr was turn off"
	_, err = models.WriteWebEvent(r, "Windows", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}
}
