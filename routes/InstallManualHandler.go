package routes

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetInstallManual(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("Authorization")

	sshKeyList, err := models.GetAllSSHKeyWithPrvKeyFromDB()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to read SSH Key from, you may try to add SSH Key again").Error())
		return
	}
	type guideLine struct {
		Line []string `json:"guide"`
	}
	if len(sshKeyList) == 0 {
		utils.ERROR(w, http.StatusBadRequest, errors.New("You have not added any SSH Key, please add one").Error())
		return
	}
	prvKey, err := models.AESDecryptKey(sshKeyList[0].PrivateKey)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to decrypt SSH Private key").Error())
		return
	}
	pubKey, err := models.GeneratePublicKey([]byte(prvKey))
	pubKeyStr := strings.Trim(string(pubKey), "\n\r")
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to generate public key from current SSH Key").Error())
		return
	}
	var guide guideLine

	guide.Line = append(guide.Line, fmt.Sprintf(`1. Open Windows CMD as Administrator then type in and run this command (Replace Server_IP to right current server ip): windows_10-agent.exe -token "%s" -key "%s" -server "{Server_ip}"`, token, pubKeyStr))
	guide.Line = append(guide.Line, "2. Choose Run Setup option")
	utils.JSON(w, http.StatusOK, guide.Line)
}
