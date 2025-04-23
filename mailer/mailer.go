package mailer

import (
	"bytes"
	"html/template"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mailer interface {
	SendSMTPMessage(templateToRender, templateName string, msg Message) error
}

type mailer struct {
	domain      string
	host        string
	port        int
	username    string
	password    string
	encryption  string
	fromAddress string
	fromName    string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
}

var _ Mailer = (*mailer)(nil)

func NewMail(domain string, host string, port int, username string, password string, encryption string, fromAddress string, fromName string) *mailer {
	return &mailer{
		domain:      domain,
		host:        host,
		port:        port,
		username:    username,
		password:    password,
		encryption:  encryption,
		fromAddress: fromAddress,
		fromName:    fromName,
	}
}

func (m *mailer) SendSMTPMessage(templateToRender, templateName string, msg Message) error {
	if msg.From == "" {
		msg.From = m.fromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.fromName
	}

	data := map[string]any{
		"message": msg.Data,
	}

	if msg.DataMap == nil {
		msg.DataMap = data
	}

	formattedMessage, err := m.buildHTMLMessage(templateToRender, templateName, msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(templateToRender, templateName, msg)
	if err != nil {
		return err
	}

	server := mail.NewSMTPClient()
	server.Host = m.host
	server.Port = m.port
	server.Username = m.username
	server.Password = m.password
	server.Encryption = m.getEncryption(m.encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To).SetSubject(msg.Subject)

	email.SetBody(mail.TextPlain, plainMessage)
	email.AddAlternative(mail.TextHTML, formattedMessage)

	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		return err
	}

	return nil
}

func (m *mailer) buildHTMLMessage(templateToRender, templateName string, msg Message) (string, error) {

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, templateName, msg.DataMap); err != nil {
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

func (m *mailer) buildPlainTextMessage(templateToRender, templateName string, msg Message) (string, error) {

	t, err := template.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, templateName, msg.DataMap); err != nil {
		return "", err
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}

func (m *mailer) inlineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}

func (m *mailer) getEncryption(s string) mail.Encryption {
	switch s {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none", "":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
