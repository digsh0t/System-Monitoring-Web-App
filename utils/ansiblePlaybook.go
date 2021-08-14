package utils

import (
	ansibler "github.com/febrianrendak/go-ansible"
)

func AnsibleFunc() {
	ansiblePlaybookConnectionOptions := &ansibler.AnsiblePlaybookConnectionOptions{
		Connection: "local",
		User:       "root",
	}

	ansiblePlaybookOptions := &ansibler.AnsiblePlaybookOptions{
		Inventory: "127.0.0.1,",
	}

	playbook := &ansibler.AnsiblePlaybookCmd{
		Playbook:          "/home/wintltr/Projects/login-api/site.yml",
		ConnectionOptions: ansiblePlaybookConnectionOptions,
		Options:           ansiblePlaybookOptions,
		ExecPrefix:        "Go-ansible example",
	}

	err := playbook.Run()
	if err != nil {
		panic(err)
	}
}
