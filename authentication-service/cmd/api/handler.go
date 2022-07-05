package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type JsonResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Authenticate struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *Config) Broker(w http.ResponseWriter, request *http.Request) {
	response := JsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	out, _ := json.MarshalIndent(response, "", "\t")

	w.Write(out)
}

func (app *Config) Authenticate(w http.ResponseWriter, request *http.Request) {
	data := Authenticate{}
	err := app.ReadJSON(w, request, &data)

	if err != nil {
		app.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}
	user, err := app.Models.User.GetByEmail(data.Email)

	if err != nil {
		app.ErrorJSON(w, errors.New("invalid credential"), http.StatusUnauthorized)
		return
	}
	isValid, err := user.PasswordMatches(data.Password)

	if err != nil || !isValid {
		app.ErrorJSON(w, errors.New("invalid credential"), http.StatusUnauthorized)
		return
	}

	err = app.logRequest(w, "Authentication", fmt.Sprintf("%s logged in", data.Email))
	if err != nil {
		app.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	payload := JSONResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logRequest(w http.ResponseWriter, name string, data string) error {
	var logPayload struct {
		Name string
		Data string
	}
	logPayload.Name = name
	logPayload.Data = data
	jsonData, _ := json.MarshalIndent(logPayload, "", "\t")

	req, err := http.NewRequest("POST", "http://logger-service/log", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return errors.New("error calling log service")
	}
	var jsonRes JSONResponse
	err = json.NewDecoder(response.Body).Decode(&jsonRes)
	if err != nil {
		return err
	}
	if jsonRes.Error {
		return errors.New(jsonRes.Message)
	}
	return nil
}
