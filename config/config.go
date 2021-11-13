package config

import (
	"encoding/base64"
	"encoding/json"

	"github.com/gorilla/securecookie"
	log "github.com/sirupsen/logrus"
)

type DbConfig struct {
	Hostname string            `json:"host"`
	Username string            `json:"user"`
	Password string            `json:"pass"`
	DbName   string            `json:"name"`
	Options  map[string]string `json:"options"`
}

//ConfigType mapping between Config and the json file that sets it
type ConfigType struct {
	MySQL DbConfig `json:"mysql"`

	Pusher PusherAPIKeyInfo `json:"pusher"`

	SecretJWT        string `json:"secret_jwt"`
	PasswordHashSalt string `json:"password_hash_salt"`
	AESKey           string `json:"aes_key"`
}

type PusherAPIKeyInfo struct {
	AppId   int    `json:"app_id"`
	Key     string `json:"key"`
	Secret  string `json:"secret"`
	Cluster string `json:"cluster"`
}

//Config exposes the application configuration storage for use in the application
var Config *ConfigType

func (conf *ConfigType) ToJSON() ([]byte, error) {
	return json.MarshalIndent(&conf, " ", "\t")
}

// LogWarning logs a warning with arbitrary field if error
func LogWarning(err error) {
	LogWarningWithFields(err, log.Fields{"level": "Warn"})
}

// LogWarningWithFields logs a warning with added field context if error
func LogWarningWithFields(err error, fields log.Fields) {
	if err != nil {
		log.WithFields(fields).Warn(err.Error())
	}
}

func (conf *ConfigType) GenerateSecrets() {
	secretByte := securecookie.GenerateRandomKey(32)
	conf.SecretJWT = base64.StdEncoding.EncodeToString(secretByte)
	secretByte = securecookie.GenerateRandomKey(32)
	conf.PasswordHashSalt = base64.StdEncoding.EncodeToString(secretByte)
	secretByte = securecookie.GenerateRandomKey(32)
	conf.AESKey = base64.StdEncoding.EncodeToString(secretByte)
}
