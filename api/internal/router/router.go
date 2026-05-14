package router

import (
	"database/sql"
	"net/http"

	"github.com/pedersenderekserif/san-diego-hackathon/api/internal/handlers"
)

func New(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	h := &handlers.Handler{DB: db}

	mux.HandleFunc("GET /v1/reporting-plans", h.ListReportingPlans)
	mux.HandleFunc("GET /v1/indexes", h.ListIndexes)
	mux.HandleFunc("GET /v1/index-templates", h.ListIndexTemplates)
	mux.HandleFunc("GET /v1/form-5500", h.ListForm5500)

	return mux
}
