package models

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Jeffail/gabs"
	vt "github.com/VirusTotal/vt-go"
	"github.com/wintltr/login-api/utils"
)

type VTReport struct {
	Attributes struct {
		ConfirmedTimeout int `json:"confirmed-timeout"`
		Failure          int `json:"failure"`
		Harmless         int `json:"harmless"`
		Malicious        int `json:"malicious"`
		Suspicious       int `json:"suspicious"`
		Timeout          int `json:"timeout"`
		TypeUnsupported  int `json:"type-unsupported"`
		Undetected       int `json:"undetected"`
	}
	URL          string `json:"url"`
	AnalysisTime string `json:"analysis_time"`
}

func ScanFile(filename string) (VTReport, error) {
	var report VTReport
	utils.EnvInit()
	apiKey := os.Getenv("VT_API_TOKEN")
	client := vt.NewClient(apiKey)

	f, err := os.Open(filename)
	if err != nil {
		return report, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return report, err
	}

	fileSHA256 := fmt.Sprintf("%x", h.Sum(nil))

	fileContent, err := os.Open("putty-64bit-0.76-installer.msi")
	if err != nil {
		return report, err
	}
	_, err = client.NewFileScanner().Scan(fileContent, "putty-64bit-0.76-installer.msi", nil)
	if err != nil {
		return report, err
	}
	file, err := client.GetObject(vt.URL("files/%s", fileSHA256))
	if err != nil {
		return report, err
	}
	out, err := json.Marshal(file)
	if err != nil {
		return report, err
	}

	//fmt.Println(string(out))
	jsonParsed, err := gabs.ParseJSON(out)
	if err != nil {
		return report, err
	}
	json.Unmarshal(jsonParsed.Path("attributes.last_analysis_stats").Bytes(), &report.Attributes)
	report.AnalysisTime = jsonParsed.Path("attributes.last_analysis_date").String()
	report.URL = strings.Trim(strings.ReplaceAll(jsonParsed.Path("links.self").String(), "https://www.virustotal.com/api/v3/files", "https://www.virustotal.com/gui/file"), `"`)
	return report, err
}

func ScanURL(url string) (VTReport, error) {
	var report VTReport
	utils.EnvInit()
	apiKey := os.Getenv("VT_API_TOKEN")
	hasher := sha256.New()
	hasher.Write([]byte(url))
	hashedURL := fmt.Sprintf("%x", hasher.Sum(nil))

	client := vt.NewClient(apiKey)
	_, err := client.NewURLScanner().Scan(url)
	if err != nil {
		return report, err
	}

	vtURL, err := client.GetObject(vt.URL("urls/%s", hashedURL))
	if err != nil {
		return report, err
	}
	out, err := json.Marshal(vtURL)
	if err != nil {
		return report, err
	}

	//fmt.Println(string(out))
	jsonParsed, err := gabs.ParseJSON(out)
	if err != nil {
		return report, err
	}
	json.Unmarshal(jsonParsed.Path("attributes.last_analysis_stats").Bytes(), &report.Attributes)
	report.AnalysisTime = jsonParsed.Path("attributes.last_analysis_date").String()
	report.URL = strings.Trim(strings.ReplaceAll(jsonParsed.Path("links.self").String(), "https://www.virustotal.com/api/v3/urls", "https://www.virustotal.com/gui/url"), `"`)
	return report, err
}
