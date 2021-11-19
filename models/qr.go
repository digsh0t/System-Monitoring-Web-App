package models

import (
	"crypto/rand"
	"encoding/base32"
	"strings"

	"github.com/dgryski/dgoogauth"
	"github.com/yeqown/go-qrcode"
)

func randStr(strSize int, randType string) string {
	var dictionary string

	if randType == "alphanum" {
		dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "alpha" {
		dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "number" {
		dictionary = "0123456789"
	}

	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}

func GenerateQR(username string) (string, string, error) {
	randomStr := randStr(20, "alphanum")

	// For Google Authenticator purpose
	// for more details see
	// https://github.com/google/google-authenticator/wiki/Key-Uri-Format
	secret := base32.StdEncoding.EncodeToString([]byte(randomStr))
	//w.Write([]byte(fmt.Sprintf("Secret : %s !\n", secret)))

	// authentication link. Remember to replace SocketLoop with yours.
	// for more details see
	// https://github.com/google/google-authenticator/wiki/Key-Uri-Format
	authLink := "otpauth://totp/lthmonitor:" + username + "?secret=" + secret + "&issuer=lthmonitor.com"

	// Encode authLink to QR codes
	// qr.H = 65% redundant level
	// see https://godoc.org/code.google.com/p/rsc/qr#Level

	qrc, err := qrcode.New(authLink)

	if err != nil {
		return "", "", err
	}

	//	var iowriter io.Writer
	tmpFilePath := "./tmp/" + randStr(6, "alphanum") + ".jpeg"

	err = qrc.Save(tmpFilePath)
	if err != nil {
		return "", "", err
	}

	// e := os.Remove(tmpFilePath)
	// if e != nil {
	// 	return nil, "", err
	// }
	return authLink, secret, nil
}

func CheckTOTP(secret string, totp string) (bool, error) {
	otpConfig := &dgoogauth.OTPConfig{
		Secret:      strings.TrimSpace(secret),
		WindowSize:  3,
		HotpCounter: 0,
	}

	trimmedToken := strings.TrimSpace(totp)

	// Validate token
	ok, err := otpConfig.Authenticate(trimmedToken)

	// if the token is invalid or expired
	if err != nil {
		return false, err
	}
	return ok, nil
}
