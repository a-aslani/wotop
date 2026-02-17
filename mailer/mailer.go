package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mailer interface {
	SendSMTPMessage(templateToRender, templateName string, msg Message) error
	SendSMTPMessageFromString(htmlContent, plainContent string, msg Message) error
	ParseString(tplString string, data map[string]any) (string, error)
	BuildHTMLMessageFromString(htmlContent string, msg Message) (string, error)
	BuildPlainTextMessageFromString(plainContent string, msg Message) (string, error)
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
	msg = m.prepareMessage(msg)

	htmlPath := fmt.Sprintf("%s.html.gohtml", templateToRender)
	plainPath := fmt.Sprintf("%s.plain.gohtml", templateToRender)

	formattedMessage, err := m.buildHTMLMessage(htmlPath, templateName, msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(plainPath, templateName, msg)
	if err != nil {
		return err
	}

	return m.send(formattedMessage, plainMessage, msg)
}

func (m *mailer) SendSMTPMessageFromString(htmlContent, plainContent string, msg Message) error {
	msg = m.prepareMessage(msg)

	formattedMessage, err := m.BuildHTMLMessageFromString(htmlContent, msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.BuildPlainTextMessageFromString(plainContent, msg)
	if err != nil {
		return err
	}

	return m.send(formattedMessage, plainMessage, msg)
}

func (m *mailer) prepareMessage(msg Message) Message {
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
	return msg
}

func (m *mailer) send(htmlBody, plainBody string, msg Message) error {
	processedSubject, err := m.ParseString(msg.Subject, msg.DataMap)
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

	fromAddress := msg.From
	if msg.FromName != "" {
		fromAddress = fmt.Sprintf("%s <%s>", msg.FromName, msg.From)
	}

	email.SetFrom(fromAddress).AddTo(msg.To).SetSubject(processedSubject)

	email.SetBody(mail.TextPlain, plainBody)
	email.AddAlternative(mail.TextHTML, htmlBody)

	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}

	return email.Send(smtpClient)
}

func (m *mailer) ParseString(tplString string, data map[string]any) (string, error) {
	t, err := template.New("inline-string").Parse(tplString)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.Execute(&tpl, data); err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func (m *mailer) buildHTMLMessage(templatePath, templateName string, msg Message) (string, error) {
	t, err := template.New("email-html").ParseFiles(templatePath)
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

func (m *mailer) buildPlainTextMessage(templatePath, templateName string, msg Message) (string, error) {
	t, err := template.New("email-plain").ParseFiles(templatePath)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, templateName, msg.DataMap); err != nil {
		return "", err
	}
	return tpl.String(), nil
}

func (m *mailer) BuildHTMLMessageFromString(htmlContent string, msg Message) (string, error) {
	t, err := template.New("email-html-string").Parse(htmlContent)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err = t.Execute(&tpl, msg.DataMap); err != nil {
		return "", err
	}
	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}
	return formattedMessage, nil
}

func (m *mailer) BuildPlainTextMessageFromString(plainContent string, msg Message) (string, error) {
	t, err := template.New("email-plain-string").Parse(plainContent)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err = t.Execute(&tpl, msg.DataMap); err != nil {
		return "", err
	}
	return tpl.String(), nil
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
