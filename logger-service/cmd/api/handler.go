package main

import (
	"net/http"
)

type JsonPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, request *http.Request) {
	var req JsonPayload
	_ = app.ReadJSON(w, request, &req)
}
