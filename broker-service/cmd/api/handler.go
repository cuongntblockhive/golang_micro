package main

import (
	"broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"net/rpc"
	"time"
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
		//app.logItem(w, requestData.Log)
		//app.logEventViaRabbit(w, requestData.Log)
		//app.logItemViaRPC(w, requestData.Log)
		app.logItemViaGRPC(w, requestData.Log)
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

func (app *Config) logEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}
	var payload JsonResponse
	payload.Error = false
	payload.Message = "Logged via RabbitMQ"

	app.WriteJSON(w, http.StatusAccepted, payload)
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logItemViaRPC(w http.ResponseWriter, l LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}
	rpcPayload := RPCPayload{
		Name: l.Name,
		Data: l.Data,
	}

	var result string
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}
	payload := JsonResponse{Error: false, Message: "Call RPC success", Data: result}
	app.WriteJSON(w, http.StatusAccepted, payload)

}

func (app *Config) pushToQueue(name string, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}
	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}
	return nil
}

func (app *Config) logItemViaGRPC(w http.ResponseWriter, l LogPayload) {
	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	defer conn.Close()

	c := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := c.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: l.Name,
			Data: l.Data,
		},
	})

	if err != nil {
		app.ErrorJSON(w, err)
		return
	}
	payload := JsonResponse{
		Error:   false,
		Message: "logged",
		Data:    res,
	}
	app.WriteJSON(w, http.StatusAccepted, payload)
}
