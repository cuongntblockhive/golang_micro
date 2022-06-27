package main

import (
	"encoding/json"
	"net/http"
)

type JsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	//Data    any    `json:"data,omitempty"`
}

func (app *Config) Broker(w http.ResponseWriter, request *http.Request) {
	response := JsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	out, _ := json.MarshalIndent(response, "", "\t")

	w.Write(out)
}
