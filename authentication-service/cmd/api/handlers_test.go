package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type RoundTripFunc func(r *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func Test_Authenticate(t *testing.T) {
	jsonPayload := map[string]interface{}{
		"email":    "test@test.com",
		"password": "verysecret",
	}

	jsonToReturnFromLoggerMock := `
	{
		"error": false,
		"message": "some message"
	}
	`
	client := NewTestClient(func(r *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(jsonToReturnFromLoggerMock)),
			Header:     make(http.Header),
		}
	})

	testApp.Client = client

	body, _ := json.Marshal(jsonPayload)
	req, _ := http.NewRequest("POST", "/authenticate", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(testApp.Authenticate)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("Expected http.StatusAccepted but got %d", rr.Code)
	}
}
