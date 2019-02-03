package response

import (
	"encoding/json"
	"net/http"
)

// JSON response
func JSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Error is a normalized data wrapper
type Error struct {
	Status string    `json:"status"`
	Error  *APIError `json:"error"`
}

// APIError Normalized struct for an API error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// APIResponseError in normalized format
func APIResponseError(w http.ResponseWriter, err APIError) {
	JSON(w, 500, Error{
		Status: "error",
		Error:  &err,
	})
}
