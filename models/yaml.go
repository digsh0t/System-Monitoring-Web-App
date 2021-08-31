package models

import ansibler "github.com/febrianrendak/go-ansible"

func RunAnsiblePlaybookWithVars(extraVars map[string]interface{}, filepath string) error {

	ansiblePlaybookOptions := &ansibler.AnsiblePlaybookOptions{
		ExtraVars: extraVars,
	}

	playbook := &ansibler.AnsiblePlaybookCmd{
		Playbook:   filepath,
		Options:    ansiblePlaybookOptions,
		ExecPrefix: "Go-ansible example",
	}

	err := playbook.Run()
	if err != nil && err.Error() != `(unreadable invalid interface type: could not find str field)` {
		panic(err)
	}
	return err
}
