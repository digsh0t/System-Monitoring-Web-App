package setup

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/wintltr/login-api/config"
)

const interactiveSetupBlurb = `
Hello! You will now be guided through a setup to:

1. Set up configuration for a MySQL/MariaDB database
2. Initialize database
3. Set up initial lthmonitor user & password
`

func askValue(prompt string, defaultValue string, item interface{}) {
	// Print prompt with optional default value
	fmt.Print(prompt)
	if len(defaultValue) != 0 {
		fmt.Print(" (default " + defaultValue + ")")
	}
	fmt.Print(": ")

	_, _ = fmt.Sscanln(defaultValue, item)

	scanErrorChecker(fmt.Scanln(item))

	// Empty line after prompt
	fmt.Println("")
}

func scanErrorChecker(n int, err error) {
	if err != nil && err.Error() != "unexpected newline" {
		log.Warn("An input error occurred: " + err.Error())
	}
}

func scanMySQL(conf *config.ConfigType) {
	askValue("DB Hostname", "127.0.0.1:3306", &conf.MySQL.Hostname)
	askValue("DB User", "root", &conf.MySQL.Username)
	askValue("DB Password", "", &conf.MySQL.Password)
	askValue("DB Name", "semaphore", &conf.MySQL.DbName)
}

func askConfirmation(prompt string, defaultValue bool, item *bool) {
	defString := "yes"
	if !defaultValue {
		defString = "no"
	}

	fmt.Print(prompt + " (yes/no) (default " + defString + "): ")

	var answer string

	scanErrorChecker(fmt.Scanln(&answer))

	switch strings.ToLower(answer) {
	case "y", "yes":
		*item = true
	case "n", "no":
		*item = false
	default:
		*item = defaultValue
	}

	// Empty line after prompt
	fmt.Println("")
}

func AskConfigConfirmation(conf *config.ConfigType) bool {
	bytes, err := conf.ToJSON()
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nGenerated configuration:\n %v\n\n", string(bytes))

	var correct bool
	askConfirmation("Is this correct?", true, &correct)
	return correct
}

func SaveConfig(conf *config.ConfigType) (configPath string) {
	configDirectory, err := os.Getwd()
	if err != nil {
		configDirectory, err = os.UserConfigDir()
		if err != nil {
			// Final fallback
			configDirectory = "./"
		}
		configDirectory = filepath.Join(configDirectory, "./")
	}

	fmt.Printf("Running: mkdir -p %v..\n", configDirectory)
	err = os.MkdirAll(configDirectory, 0755)
	if err != nil {
		log.Panic("Could not create config directory: " + err.Error())
	}

	// Marshal config to json
	bytes, err := conf.ToJSON()
	if err != nil {
		panic(err)
	}

	configPath = filepath.Join(configDirectory, "config.json")
	if err = ioutil.WriteFile(configPath, bytes, 0644); err != nil {
		panic(err)
	}

	fmt.Printf("Configuration written to %v..\n", configPath)
	return
}

func InteractiveSetup(conf *config.ConfigType) {
	fmt.Print(interactiveSetupBlurb)

	dbPrompt := `What database to use:
   1 - MySQL
   2 - ...
`

	var db int
	askValue(dbPrompt, "1", &db)

	switch db {
	case 1:
		scanMySQL(conf)
	}

	askValue("Pusher App ID (notification extension,you can get it from pusher.com)", "", &conf.Pusher.AppId)
	askValue("Pusher API Secret ", "", &conf.Pusher.Secret)
	askValue("Pusher API Key", "", &conf.Pusher.Key)
	askValue("Pusher API Cluster", "", &conf.Pusher.Cluster)
}
