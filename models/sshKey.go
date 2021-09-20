package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/utils"
)

type SSHKey struct {
	SSHKeyId   int    `json:"sshKeyId"`
	KeyName    string `json:"sshKeyName"`
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

func AESDecryptKey(encryptedPrivateKey string) (string, error) {
	utils.EnvInit()
	aesKey := os.Getenv("AES_KEY")
	plaintext, err := decrypt([]byte(aesKey), encryptedPrivateKey)
	if err != nil {
		return plaintext, errors.New("fail to decrypt message")
	}
	return plaintext, err
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
	defer stmt.Close()
	_, err = stmt.Exec(sshKey.KeyName, sshKey.PrivateKey, sshKey.CreatorId)
	if err != nil {
		return false, err
	}
	return true, err
}

func GetAllSSHKeyFromDB() ([]SSHKey, error) {
	var sshKey SSHKey
	var sshKeyList []SSHKey

	db := database.ConnectDB()
	defer db.Close()
	rows, err := db.Query("SELECT sk_key_id, sk_key_name, creator_id FROM ssh_keys")
	if err != nil {
		return sshKeyList, err
	}
	for rows.Next() {
		err = rows.Scan(&sshKey.SSHKeyId, &sshKey.KeyName, &sshKey.CreatorId)
		if err != nil {
			return sshKeyList, err
		}
		sshKeyList = append(sshKeyList, sshKey)
	}
	return sshKeyList, err
}

func GetSSHKeyFromId(id int) (SSHKey, error) {
	var returnedSSHKey SSHKey

	db := database.ConnectDB()
	defer db.Close()
	row := db.QueryRow("SELECT sk_key_id, sk_key_name, sk_private_key, creator_id FROM ssh_keys WHERE sk_key_id= ?", id)

	err := row.Scan(&returnedSSHKey.SSHKeyId, &returnedSSHKey.KeyName, &returnedSSHKey.PrivateKey, &returnedSSHKey.CreatorId)
	if err != nil {
		return returnedSSHKey, err
	}
	return returnedSSHKey, err
}

//Generate SSH public key from private key
func GeneratePublicKey(privatekey []byte) ([]byte, error) {
	priv, err := ssh.ParsePrivateKey(privatekey)
	if err != nil {
		utils.EnvInit()
		priv, err = ssh.ParsePrivateKeyWithPassphrase(privatekey, []byte(os.Getenv("SECRET_SSH_PASSPHRASE")))
		if err != nil {
			fmt.Printf("Error while parsing Private key: %s", err)
		}
	}
	publicKey := priv.PublicKey()

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicKey)

	return pubKeyBytes, nil
}

func SSHKeyDelete(id int) (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM ssh_keys WHERE sk_key_id = ?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return false, errors.New("no SSH Connections with this ID exists")
	}

	return true, err
}

func GetAllSSHKeyWithPrvKeyFromDB() ([]SSHKey, error) {
	var sshKey SSHKey
	var sshKeyList []SSHKey

	db := database.ConnectDB()
	defer db.Close()
	rows, err := db.Query("SELECT sk_key_id, sk_key_name, sk_private_key, creator_id FROM ssh_keys")
	if err != nil {
		return sshKeyList, err
	}
	for rows.Next() {
		err = rows.Scan(&sshKey.SSHKeyId, &sshKey.KeyName, &sshKey.PrivateKey, &sshKey.CreatorId)
		if err != nil {
			return sshKeyList, err
		}
		sshKeyList = append(sshKeyList, sshKey)
	}
	return sshKeyList, err
}

func GetKeyIdFromPublicKey(pubKey string) (int, error) {
	keyList, err := GetAllSSHKeyWithPrvKeyFromDB()
	var privKey string
	if err != nil {
		return -1, err
	}
	for _, key := range keyList {
		privKey, err = AESDecryptKey(key.PrivateKey)
		if err != nil {
			return -1, err
		}
		currentPub, err := GeneratePublicKey([]byte(privKey))
		if err != nil {
			return -1, err
		}
		if pubKey+"\n" == string(currentPub) {
			return key.SSHKeyId, err
		}
	}
	return -1, err
}
