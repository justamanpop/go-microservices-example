package main

import (
	"net/http"
)

func (cfg *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	type JsonRequest struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var requestPayload JsonRequest

	err := cfg.readJson(w, r, &requestPayload)
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	msg := Message{
		From:    requestPayload.From,
		To:      requestPayload.To,
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}

	err = cfg.Mailer.SendSMTPMessage(msg)
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	responseJson := jsonResponse{
		Error:   false,
		Message: "Mail sent to " + requestPayload.To,
	}
	cfg.writeJson(w, http.StatusAccepted, responseJson)
}
