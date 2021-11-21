package routes

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetInstallManual(w http.ResponseWriter, r *http.Request) {

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

	token := r.Header.Get("Authorization")

	sshKeyList, err := models.GetAllSSHKeyWithPrvKeyFromDB()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read SSH Key from, you may try to add SSH Key again").Error())
		return
	}
	type guideLine struct {
		Line []string `json:"guide"`
	}
	if len(sshKeyList) == 0 {
		utils.ERROR(w, http.StatusBadRequest, errors.New("you have not added any SSH Key, please add one").Error())
		return
	}
	prvKey, err := models.AESDecryptKey(sshKeyList[0].PrivateKey)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to decrypt SSH Private key").Error())
		return
	}
	pubKey, err := models.GeneratePublicKey([]byte(prvKey))
	pubKeyStr := strings.Trim(string(pubKey), "\n\r")
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to generate public key from current SSH Key").Error())
		return
	}
	var guide guideLine

	guide.Line = append(guide.Line, fmt.Sprintf(`1. Open Windows CMD as Administrator then type in and run this command (Replace Server_IP to right current server ip): windows_10-agent.exe -token "%s" -key "%s" -server "{Server_ip}"`, token, pubKeyStr))
	guide.Line = append(guide.Line, "2. Choose Run Setup option")
	utils.JSON(w, http.StatusOK, guide.Line)
}
