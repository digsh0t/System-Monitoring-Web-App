package models

import (
	"errors"
	"io"
	"net/http"
	"text/template"
	"time"

	"github.com/wintltr/login-api/auth"
	"gopkg.in/gomail.v2"
)

type SmtpInfo struct {
	EmailSender   string `json:"email_sender"`
	EmailPassword string `json:"email_password"`
	SMTPHost      string `json:"smtp_host"`
	SMTPPort      string `json:"smtp_port"`
}

func (sI SmtpInfo) SendReportMail(filepath string, receiver []string, ccer []string, bccer []string, r *http.Request) error {
	subject := "LTH Monitor Report requestes by user at " + time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
	m := gomail.NewMessage()
	if receiver == nil {
		return errors.New("Please enter the receiver email")
	}
	m.SetHeader("From", sI.EmailSender)
	m.SetHeader("To", receiver...)
	if ccer != nil {
		m.SetHeader("Cc", ccer...)
	}
	if bccer != nil {
		m.SetHeader("Bcc", bccer...)
	}
	m.SetHeader("Subject", subject)
	t, _ := template.ParseFiles("./template/mail_template.html")

	// Get Current User Web app
	tokenData, err := auth.RetrieveTokenData(r)
	if err != nil {
		return err
	}
	userWebApp := tokenData.Username
	m.AddAlternativeWriter("text/html", func(w io.Writer) error {
		return t.Execute(w, struct {
			Username string
		}{

			Username: userWebApp,
		})
	})

	m.Attach(filepath)

	d := gomail.NewDialer("smtp.gmail.com", 587, sI.EmailSender, sI.EmailPassword)

	err = d.DialAndSend(m)
	return err
}
