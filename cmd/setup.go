package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wintltr/login-api/config"
	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/setup"
)

func init() {
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Perform interactive setup",
	Run: func(cmd *cobra.Command, args []string) {
		doSetup()
	},
}

func readNewline(pre string, stdin *bufio.Reader) string {
	fmt.Print(pre)

	str, err := stdin.ReadString('\n')
	config.LogWarning(err)
	str = strings.Replace(strings.Replace(str, "\n", "", -1), "\r", "", -1)

	return str
}

func doSetup() int {
	var conf *config.ConfigType
	for {
		conf = &config.ConfigType{}
		conf.GenerateSecrets()
		setup.InteractiveSetup(conf)

		if setup.AskConfigConfirmation(conf) {
			break
		}

		fmt.Println()
	}
	databae, err := database.TestConnectionMysqlDB(*conf)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("success")
	}
	err = database.MigrateDB(databae, "./database/Capstone_WA_30-08-2021.sql")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	setup.SaveConfig(conf)

	// store := factory.CreateStore()
	// if err := store.Connect(); err != nil {
	// 	fmt.Printf("Cannot connect to database!\n %v\n", err.Error())
	// 	os.Exit(1)
	// }

	// fmt.Println("Running DB Migrations..")
	// if err := store.Migrate(); err != nil {
	// 	fmt.Printf("Database migrations failed!\n %v\n", err.Error())
	// 	os.Exit(1)
	// }

	stdin := bufio.NewReader(os.Stdin)

	rsyslogConfigPath := readNewline("\n\n > Rsyslog config path (default: /etc/rsyslog.conf): ", stdin)
	if rsyslogConfigPath == "" {
		rsyslogConfigPath = "/etc/rsyslog.conf"
	}
	models.SetupRsyslogServer(rsyslogConfigPath)

	var user models.User
	user.Username = readNewline("\n\n > Username: ", stdin)
	user.Username = strings.ToLower(user.Username)
	user.Password = readNewline(" > Password: ", stdin)
	user.Name = readNewline(" > Realname: ", stdin)
	user.Role = "admin"
	ok, err := models.AddWebAppUser(user)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if ok {
		fmt.Printf("\n You are all setup %v!\n", user.Name)
	}

	// util.LogWarning(err)

	// if existingUser.ID > 0 {
	// 	// user already exists
	// 	fmt.Printf("\n Welcome back, %v! (a user with this username/email is already set up..)\n\n", existingUser.Name)
	// } else {
	// 	user.Name = readNewline(" > Your name: ", stdin)
	// 	user.Pwd = readNewline(" > Password: ", stdin)
	// 	user.Admin = true

	// 	if _, err := store.CreateUser(user); err != nil {
	// 		fmt.Printf(" Inserting user failed. If you already have a user, you can disregard this error.\n %v\n", err.Error())
	// 		os.Exit(1)
	// 	}

	// 	fmt.Printf("\n You are all setup %v!\n", user.Name)
	// }

	fmt.Printf(" Re-launch this program by running\n\n./lthmonitor service\n\n")
	fmt.Printf(" To run as daemon:\n\nnohup ./lthmonitor &\n\n")
	fmt.Printf(" You can login with username: %v.\n", user.Username)

	return 0
}
