package main

import (
	"log"
	"net/http"

	"github.com/justamanpop/microservices_logger/data"
)

type JsonPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (cfg *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	var reqPayload JsonPayload
	cfg.readJson(w, r, &reqPayload)

	logToInsert := data.LogEntry{
		Name: reqPayload.Name,
		Data: reqPayload.Data,
	}

	err := cfg.Models.LogEntry.Insert(logToInsert)
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	response := jsonResponse{
		Error:   false,
		Message: "Successfully logged",
	}

	data, err := logToInsert.All()
	for index, entry := range data {
		log.Printf("all the data at index %d is %+v", index, *entry)
	}
	cfg.writeJson(w, http.StatusAccepted, response)
}
