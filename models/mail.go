package models

import (
	"io"
	"text/template"
	"time"

	"gopkg.in/gomail.v2"
)

type SmtpInfo struct {
	EmailSender   string `json:"email_sender"`
	EmailPassword string `json:"email_password"`
	SMTPHost      string `json:"smtp_host"`
	SMTPPort      string `json:"smtp_port"`
}

func (sI SmtpInfo) SendReportMail(filepath string, receiver []string) {
	subject := "LTH Monitor Report requestes by user at " + time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
	m := gomail.NewMessage()
	m.SetHeader("From", sI.EmailSender)
	m.SetHeader("To", receiver...)
	m.SetHeader("Subject", subject)
	t, _ := template.ParseFiles("./template/mail_template.html")
	m.AddAlternativeWriter("text/html", func(w io.Writer) error {
		return t.Execute(w, struct {
			Username string
		}{
			Username: "Le Xuan Tri",
		})
	})

	m.Attach(filepath)

	d := gomail.NewDialer("smtp.gmail.com", 587, sI.EmailSender, sI.EmailPassword)

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}
