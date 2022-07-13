package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type JsonResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type RequestSubmissionPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (app *Config) Broker(w http.ResponseWriter, request *http.Request) {
	response := JsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	out, _ := json.MarshalIndent(response, "", "\t")

	w.Write(out)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, request *http.Request) {
	requestData := RequestSubmissionPayload{}
	err := app.ReadJSON(w, request, &requestData)
	if err != nil {
		app.ErrorJSON(w, err)
	}
	switch requestData.Action {
	case "auth":
		app.authenticate(w, requestData.Auth)
	case "log":
		app.logItem(w, requestData.Log)
	case "mail":
		app.sendMail(w, requestData.Mail)
	default:
		app.ErrorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, authPayload AuthPayload) {
	jsonData, _ := json.MarshalIndent(authPayload, "", "\t")

	req, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		app.ErrorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.ErrorJSON(w, errors.New("error calling auth service"))
		return
	}

	var jsonRes JSONResponse
	err = json.NewDecoder(response.Body).Decode(&jsonRes)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	if jsonRes.Error == true {
		app.ErrorJSON(w, err, http.StatusUnauthorized)
		return
	}
	res := JSONResponse{
		Error:   false,
		Message: "Authenticated",
		Data:    jsonRes.Data,
	}
	app.WriteJSON(w, http.StatusAccepted, res)
}

func (app *Config) logItem(w http.ResponseWriter, logPayload LogPayload) {
	jsonData, _ := json.MarshalIndent(logPayload, "", "\t")

	req, err := http.NewRequest("POST", "http://logger-service/log", bytes.NewBuffer(jsonData))
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.ErrorJSON(w, errors.New("error calling log service"))
		return
	}

	var jsonRes JSONResponse
	err = json.NewDecoder(response.Body).Decode(&jsonRes)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	if jsonRes.Error {
		app.ErrorJSON(w, errors.New(jsonRes.Message), http.StatusBadRequest)
		return
	}
	res := JSONResponse{
		Error:   false,
		Message: "Logged",
		Data:    jsonRes.Data,
	}
	app.WriteJSON(w, http.StatusAccepted, res)
}

func (app *Config) sendMail(w http.ResponseWriter, mailPayload MailPayload) {
	jsonData, _ := json.MarshalIndent(mailPayload, "", "\t")

	mailUrl := "http://mail-service/send"
	req, err := http.NewRequest("POST", mailUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.ErrorJSON(w, errors.New("error calling mail service"))
		return
	}
	res := JSONResponse{
		Error:   false,
		Message: "Message sent to " + mailPayload.To,
	}
	app.WriteJSON(w, http.StatusAccepted, res)
}
