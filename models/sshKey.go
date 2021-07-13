package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/utils"
)

type SSHKey struct {
	SSHKeyId   int    `json:"sshKeyId"`
	KeyName    string `json:"keyName"`
	PrivateKey string `json:"privateKey"`
	CreatorId  int    `json:"creatorId"`
}

func AESEncryptKey(privateKey string) string {
	utils.EnvInit()
	aesKey := os.Getenv("AES_KEY")
	ciphertext, err := encrypt([]byte(aesKey), privateKey)
	if err != nil {
		fmt.Println(err)
	}
	return ciphertext
}

func encrypt(key []byte, message string) (encmess string, err error) {
	plainText := []byte(message)

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	//returns to base64 encoded string
	encmess = base64.URLEncoding.EncodeToString(cipherText)
	return
}

func decrypt(key []byte, securemess string) (decodedmess string, err error) {
	cipherText, err := base64.URLEncoding.DecodeString(securemess)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		fmt.Println("Ciphertext block size is too short!")
		return
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	decodedmess = string(cipherText)
	return
}

func (sshKey *SSHKey) InsertSSHKeyToDB() (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO ssh_keys (sk_key_name, sk_private_key, creator_id) VALUES (?,?,?)")
	if err != nil {
		return false, err
	}
	_, err = stmt.Exec(sshKey.KeyName, sshKey.PrivateKey, sshKey.CreatorId)
	if err != nil {
		return false, err
	}
	return true, err
}
