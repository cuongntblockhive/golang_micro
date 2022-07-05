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
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
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
	request_data := RequestSubmissionPayload{}
	err := app.ReadJSON(w, request, &request_data)
	if err != nil {
		app.ErrorJSON(w, err)
	}
	switch request_data.Action {
	case "auth":
		app.authenticate(w, request_data.Auth)
	case "log":
		app.logItem(w, request_data.Log)
	default:
		app.ErrorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, authPayload AuthPayload) {
	json_data, _ := json.MarshalIndent(authPayload, "", "\t")

	req, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(json_data))
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
	payload := JSONResponse{
		Error:   false,
		Message: "Authenticated",
		Data:    jsonRes.Data,
	}
	app.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logItem(w http.ResponseWriter, logPayload LogPayload) {
	json_data, _ := json.MarshalIndent(logPayload, "", "\t")

	req, err := http.NewRequest("POST", "http://logger-service/log", bytes.NewBuffer(json_data))
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
	payload := JSONResponse{
		Error:   false,
		Message: "Logged",
		Data:    jsonRes.Data,
	}
	app.WriteJSON(w, http.StatusAccepted, payload)
}
