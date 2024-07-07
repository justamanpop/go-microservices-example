package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/rpc"
	"time"

	"github.com/justamanpop/microservices_broker/event"
	"github.com/justamanpop/microservices_broker/logs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"` //we will have on field per action, like one for logging, one for auth, etc.
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

func (cfg *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Broker here",
	}

	_ = cfg.writeJson(w, http.StatusOK, payload)
}

func (cfg *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := cfg.readJson(w, r, &requestPayload)
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		cfg.authenticate(w, requestPayload.Auth)
	case "log":
		// cfg.logItem(w, requestPayload.Log)
		// cfg.logEventViaRabbit(w, requestPayload.Log)
		cfg.logEventViaRpc(w, requestPayload.Log)
	case "mail":
		cfg.mail(w, requestPayload.Mail)
	default:
		cfg.errorJson(w, errors.New("unsupported action"))
	}
}

func (cfg *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	//create a JSON to send to auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")
	//docker compose takes care of resolving the URL here
	req, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		cfg.errorJson(w, errors.New("invalid credentials"))
		return
	} else if res.StatusCode != http.StatusAccepted {
		cfg.errorJson(w, errors.New("error calling auth service"))
		return
	}

	var authRes jsonResponse

	err = json.NewDecoder(res.Body).Decode(&authRes)
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	if authRes.Error {
		cfg.errorJson(w, errors.New(authRes.Message), http.StatusUnauthorized)
		return
	}

	//based on auth service response, return response from broker
	var responsePayload jsonResponse
	responsePayload.Error = false
	responsePayload.Message = "Authenticated"
	responsePayload.Data = authRes.Data

	cfg.writeJson(w, http.StatusAccepted, responsePayload)
}

func (cfg *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	//docker compose takes care of resolving the URL here
	req, err := http.NewRequest("POST", "http://logger-service/log", bytes.NewBuffer(jsonData))
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		cfg.errorJson(w, errors.New("error calling logger service"))
		return
	}

	var logRes jsonResponse

	err = json.NewDecoder(res.Body).Decode(&logRes)
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	if logRes.Error {
		cfg.errorJson(w, errors.New(logRes.Message))
		return
	}

	//based on auth service response, return response from broker
	var responsePayload jsonResponse
	responsePayload.Error = false
	responsePayload.Message = "Message logged"
	responsePayload.Data = logRes.Data

	cfg.writeJson(w, http.StatusAccepted, responsePayload)
}

func (cfg *Config) mail(w http.ResponseWriter, m MailPayload) {
	jsonData, _ := json.MarshalIndent(m, "", "")

	req, err := http.NewRequest("POST", "http://mailer-service/send", bytes.NewBuffer(jsonData))
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	defer res.Body.Close()

	var mailRes jsonResponse

	err = json.NewDecoder(res.Body).Decode(&mailRes)
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	if res.StatusCode != http.StatusAccepted {
		cfg.errorJson(w, errors.New("error calling mail service"))
		return
	}

	if mailRes.Error {
		cfg.errorJson(w, errors.New(mailRes.Message))
		return
	}

	//based on auth service response, return response from broker
	var responsePayload jsonResponse
	responsePayload.Error = false
	responsePayload.Message = "Mail sent from " + m.From + " to " + m.To
	responsePayload.Data = mailRes.Data

	cfg.writeJson(w, http.StatusAccepted, responsePayload)
}

func (cfg *Config) logEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := cfg.pushToQueue(l)
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	var responsePayload jsonResponse
	responsePayload.Error = false
	responsePayload.Message = "Pushed log event to RabbitMQ queue"
	cfg.writeJson(w, http.StatusAccepted, responsePayload)
}

func (cfg *Config) pushToQueue(l LogPayload) error {
	emitter, err := event.NewEmitter(cfg.RabbitConn)
	if err != nil {
		return err
	}

	jsonToEmit, _ := json.MarshalIndent(&l, "", "\t")
	err = emitter.Emit(string(jsonToEmit), "log.INFO")
	if err != nil {
		return err
	}
	return err
}

type RpcPayload struct {
	Name string
	Data string
}

func (cfg *Config) logEventViaRpc(w http.ResponseWriter, l LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		cfg.errorJson(w, err)
		return
	}
	rpcPayload := RpcPayload{
		Name: l.Name,
		Data: l.Data,
	}

	var res string

	err = client.Call("RpcServer.Log", rpcPayload, &res)
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	responseJson := jsonResponse{
		Error:   false,
		Message: res,
	}

	cfg.writeJson(w, http.StatusAccepted, responseJson)
}

func (cfg *Config) logViaGrpc(w http.ResponseWriter, r *http.Request) {
	var reqPayload RequestPayload
	err := cfg.readJson(w, r, &reqPayload)
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		cfg.errorJson(w, err)
		return
	}
	defer conn.Close()

	c := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: reqPayload.Log.Name,
			Data: reqPayload.Log.Data,
		},
	})
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	var responsePayload jsonResponse
	responsePayload.Error = false
	responsePayload.Message = "Logged via grpc"
	cfg.writeJson(w, http.StatusAccepted, responsePayload)
}
