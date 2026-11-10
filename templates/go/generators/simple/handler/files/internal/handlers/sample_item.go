// Package handlers contém os handlers HTTP.
package handlers

import (
	"encoding/json"
	"net/http"
)

// ListSampleItems responde com a lista de sample-items.
func ListSampleItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]string{})
}
