package main

import (
	"bytes"
	template2 "html/template"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	UserName    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        interface{}
	DataMap     map[string]interface{}
}

func (m *Mail) SendSMTPMessage(msg Message) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}
	data := map[string]interface{}{
		"message": msg.Data,
	}
	msg.DataMap = data

	fmtMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	template := "template/mail.html.tmpl"
	t, err := template2.New("email-html").ParseFiles(template)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	fmtMessage := tpl.String()
	fmtMessage, err = m.inlineCSS(fmtMessage)
	if err != nil {
		return "", err
	}
	return fmtMessage, nil
}

func (m *Mail) inlineCSS(s string) (string, error) {

}
