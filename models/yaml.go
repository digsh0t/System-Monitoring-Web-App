package models

type YamlInfo struct {
	Host     string `json:"host"`
	FileName string `json:"filename"`
	Mode     string `json:"mode"`
	Package  string `json:"package"`
}
