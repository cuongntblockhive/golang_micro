package main

import (
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

	payload := JSONResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.WriteJSON(w, http.StatusAccepted, payload)
}
