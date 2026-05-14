package router

import (
	"database/sql"
	"net/http"

	"github.com/pedersenderekserif/san-diego-hackathon/api/internal/handlers"
)

func New(db *sql.DB, aetnaSource *handlers.AetnaSource) http.Handler {
	mux := http.NewServeMux()

	h := &handlers.Handler{DB: db, AetnaSource: aetnaSource}

	mux.HandleFunc("GET /v1/reporting-plans/filters", h.GetReportingPlanFilters)
	mux.HandleFunc("GET /v1/reporting-plans", h.ListReportingPlans)
	mux.HandleFunc("GET /v1/indexes", h.ListIndexes)
	mux.HandleFunc("GET /v1/index-templates/payors", h.GetIndexTemplatePayors)
	mux.HandleFunc("GET /v1/index-templates", h.ListIndexTemplates)
	mux.HandleFunc("GET /v1/form-5500", h.ListForm5500)
	mux.HandleFunc("GET /v1/aetna-mrf/plans", h.GetAetnaMRFPlans)

	return mux
}
