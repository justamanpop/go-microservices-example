package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

func (cfg *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var reqPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := cfg.readJson(w, r, &reqPayload)
	if err != nil {
		cfg.errorJson(w, err, http.StatusBadRequest)
		return
	}

	user, err := cfg.Repo.GetByEmail(reqPayload.Email)
	if err != nil {
		cfg.errorJson(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	valid, err := cfg.Repo.PasswordMatches(reqPayload.Password, *user)
	if err != nil || !valid {
		cfg.errorJson(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	err = cfg.logRequest("Login", fmt.Sprintf("User with email %s logged in", reqPayload.Email))
	if err != nil {
		cfg.errorJson(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in as user %s", user.Email),
		Data:    user,
	}

	cfg.writeJson(w, http.StatusAccepted, payload)

}

func (cfg *Config) logRequest(name, data string) error {
	var logPayload struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	logPayload.Name = name
	logPayload.Data = data

	jsonData, _ := json.MarshalIndent(logPayload, "", "\t")

	req, err := http.NewRequest("POST", "http://logger-service/log", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := cfg.Client.Do(req)
	var logRes jsonResponse

	err = json.NewDecoder(res.Body).Decode(&logRes)
	log.Printf("log resposne from auth service is %+v", logRes)

	if err != nil {
		return err
	}

	return nil

}
