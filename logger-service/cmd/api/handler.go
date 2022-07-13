package main

import (
	"logger/data"
	"net/http"
)

type JsonResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type JsonPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, request *http.Request) {
	var req JsonPayload
	err := app.ReadJSON(w, request, &req)
	if err != nil {
		app.ErrorJSON(w, err)
	}
	log := data.LogEntry{
		Name: req.Name,
		Data: req.Data,
	}

	err = app.Models.LogEntry.Insert(log)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	res := JsonResponse{
		Error:   false,
		Message: "logged",
		Data:    log,
	}
	app.WriteJSON(w, http.StatusAccepted, res)
}
