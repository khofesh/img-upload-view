package reqres

import (
	"encoding/json"
	"net/http"
)

// Define a writeJSON() helper for sending responses.
func WriteJSON(w http.ResponseWriter, status int, data map[string]any, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}
