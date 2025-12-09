package mailer

import (
	"bytes"
	"embed"
	"time"

	"github.com/wneessen/go-mail"

	ht "html/template"
	tt "text/template"
)

//go:embed templates/*
var templateFS embed.FS

type Mailer struct {
	client *mail.Client
	sender string
}

func New(host string, port int, username, password, sender string) (*Mailer, error) {
	client, err := mail.NewClient(
		host,
		mail.WithSMTPAuth(mail.SMTPAuthLogin),
		mail.WithPort(port),
		mail.WithUsername(username),
		mail.WithPassword(password),
		mail.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, err
	}
	mailer := &Mailer{
		client: client,
		sender: sender,
	}
	return mailer, nil
}

func (m *Mailer) Send(recipient string, templateFile string, data any) error {
	textTmpl, err := tt.New("").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	htmlTmpl, err := ht.New("").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// Build subject, plain text, and HTML bodies from the template.
	subject := new(bytes.Buffer)
	if err = htmlTmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	if err = textTmpl.ExecuteTemplate(plainBody, "plainBody", data); err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = htmlTmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMsg()

	err = msg.To(recipient)
	if err != nil {
		return err
	}

	err = msg.From(m.sender)
	if err != nil {
		return err
	}

	msg.Subject(subject.String())
	msg.SetBodyString(mail.TypeTextPlain, plainBody.String())
	msg.AddAlternativeString(mail.TypeTextHTML, htmlBody.String())

	// DialAndSend will establish a connection for this message and close it afterwards.
	return m.client.DialAndSend(msg)
}
