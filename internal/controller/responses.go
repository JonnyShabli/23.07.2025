package controller

import (
	"encoding/json"
	"net/http"

	"github.com/JonnyShabli/23.07.2025/pkg/logster"
)

type Response struct {
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func SuccessResponse(w http.ResponseWriter, log logster.Logger, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if data == nil {
		data = ""
	}
	resp := Response{
		Msg:  msg,
		Data: data,
	}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		log.WithError(err).Errorf("json encode error")
		return
	}
}

func BadRequestResponse(w http.ResponseWriter, log logster.Logger, errMsg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if data == nil {
		data = ""
	}
	resp := Response{
		Msg:  errMsg,
		Data: data,
	}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		log.WithError(err).Errorf("json encode error")
		return
	}

}

func ErrorResponse(w http.ResponseWriter, log logster.Logger, errMsg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	if data == nil {
		data = ""
	}
	resp := Response{
		Msg:  errMsg,
		Data: data,
	}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		log.WithError(err).Errorf("json encode error")
		return
	}

}
