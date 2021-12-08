package models

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

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
	var counter int = 0
	url := "https://www.virustotal.com/api/v3/files"
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

	values := map[string]io.Reader{
		"file": mustOpen(filename), // lets assume its this file
	}
	err = Upload(http.DefaultClient, url, values, f.Name())
	if err != nil {
		return report, err
	}

	report, err = getFileVTReport(client, fileSHA256)
	if err != nil {
		return report, err
	}
	for report.AnalysisTime == "{}" {
		report, err = getFileVTReport(client, fileSHA256)
		if err != nil {
			return report, err
		}
		if report.AnalysisTime == "{}" {
			if counter == 1 {
				return report, errors.New("file is in queue to scan, please try later")
			}
			time.Sleep(30 * time.Second)
			counter++
		}
	}
	return report, err
}

func Upload(client *http.Client, url string, values map[string]io.Reader, filename string) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	contentType := getContentTypeFromFile(filename)
	newBuffer := strings.ReplaceAll(b.String(), "Content-Type: application/octet-stream", "Content-Type: "+contentType)
	//b.Bytes() = []byte(newBuffer)
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(newBuffer))
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Add("Accept", "application/json")
	req.Header.Add("x-apikey", os.Getenv("VT_API_TOKEN"))
	req.Header.Add("Content-Type", "multipart/form-data")
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

func getContentTypeFromFile(filename string) string {
	var extension string
	contentTypes := map[string]string{
		"msi":   "application/x-msi",
		"octet": "application/octet-stream",
	}
	tmp := strings.Split(filename, ".")
	if tmp == nil {
		extension = "octet"
	} else {
		extension = tmp[len(tmp)-1]
	}
	return contentTypes[extension]
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}

func getFileVTReport(client *vt.Client, fileSHA256 string) (VTReport, error) {
	var report VTReport
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
	var counter int
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

	report, err = getURLVTReport(client, hashedURL)
	if err != nil {
		return report, err
	}
	for report.AnalysisTime == "{}" {
		report, err = getURLVTReport(client, hashedURL)
		if err != nil {
			return report, err
		}
		if report.AnalysisTime == "{}" {
			if counter == 1 {
				return report, errors.New("file is in queue to scan, please try later")
			}
			time.Sleep(30 * time.Second)
			counter++
		}
	}
	return report, err
}

func getURLVTReport(client *vt.Client, hashedURL string) (VTReport, error) {
	var report VTReport
	file, err := client.GetObject(vt.URL("urls/%s", hashedURL))
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
	report.URL = strings.Trim(strings.ReplaceAll(jsonParsed.Path("links.self").String(), "https://www.virustotal.com/api/v3/urls", "https://www.virustotal.com/gui/url"), `"`)
	return report, err
}
