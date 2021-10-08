package apimodel

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type ErrorMessage struct {
	ErrStatusCode int    `json:"status_code"`
	ErrMessage    string `json:"message"`
}

func (e *ErrorMessage) StatusCode() int {
	return e.ErrStatusCode
}

func (e *ErrorMessage) Title() string {
	return e.ErrMessage
}

func (e *ErrorMessage) Error() string {
	if e.ErrMessage != "" {
		return strconv.Itoa(e.ErrStatusCode) + ":" + e.ErrMessage
	} else {
		return strconv.Itoa(e.ErrStatusCode)
	}
}

func (v ErrorMessage) SendError(w http.ResponseWriter) {
	message := v.ErrMessage
	if message == "" {
		switch v.ErrStatusCode {
		case http.StatusNotFound:
			message = "Page not found"
		case http.StatusForbidden:
			message = "Forbidden"
		case http.StatusServiceUnavailable:
			message = "Service unavailable"
		case http.StatusBadRequest:
			message = "Bad request"
		default:
			message = "Internal error"
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(v.ErrStatusCode)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		logrus.Panicf("error when encoding error: %v", err)
	}
}

//errors message
var WrongParametersErrorMessage = ErrorMessage{
	ErrStatusCode: http.StatusBadRequest,
	ErrMessage:    "unable to parse parameters",
}
