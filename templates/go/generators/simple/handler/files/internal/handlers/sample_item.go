// Package handlers holds the HTTP handlers.
package handlers

import (
	"encoding/json"
	"net/http"
)

// ListSampleItems responds with the list of sample-items.
func ListSampleItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]string{})
}
