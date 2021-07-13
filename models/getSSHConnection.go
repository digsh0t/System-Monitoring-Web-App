package models

type GetSSHConnectionInfo struct {
	Sc_username string `json:"sc_username"`
	Sc_host     string `json:"sc_host"`
	Sc_port     int    `json:"sc_port"`
	Sk_key_name string `json:"sk_key_name"`
}
