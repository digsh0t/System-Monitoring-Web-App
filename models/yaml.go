package models

type YamlInfo struct {
	Host    []string `json:"host"`
	File    string   `json:"file"`
	Mode    string   `json:"mode"`
	Package string   `json:"package"`
	Link    string   `json:"link"`
}
