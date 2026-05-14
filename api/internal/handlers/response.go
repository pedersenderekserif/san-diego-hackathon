package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type Handler struct {
	DB          *sql.DB
	AetnaSource *AetnaSource
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
