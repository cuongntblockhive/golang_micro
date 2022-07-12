package main

import "net/http"

func (app *Config) SendMail(w http.ResponseWriter, request *http.Request) {
	type MailMsg struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}
	var payload MailMsg
	err := app.ReadJSON(w, request, &payload)
	if err != nil {
		app.ErrorJSON(w, err)
	}
	msg := Message{
		From:    payload.From,
		To:      payload.To,
		Subject: payload.Subject,
		Data:    payload.Message,
	}
	err = app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		app.ErrorJSON(w, err)
	}

	response := JSONResponse{
		Error:   false,
		Message: "Sent to: " + payload.To}
	app.WriteJSON(w, http.StatusAccepted, response)
}
