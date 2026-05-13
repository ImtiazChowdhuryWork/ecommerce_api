package utils

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func WriteSuccess(w http.ResponseWriter, status int, data interface{}) {
	WriteJSON(w, status, Response{Success: true, Data: data})
}

func WriteMessage(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, Response{Success: true, Message: message})
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, Response{Success: false, Error: message})
}
