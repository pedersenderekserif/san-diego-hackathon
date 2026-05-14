package router

import (
	"database/sql"
	"net/http"

	"github.com/pedersenderekserif/san-diego-hackathon/api/internal/handlers"
)

func New(db *sql.DB, aetnaSource *handlers.AetnaSource, bcbsilSource *handlers.BCBSILSource, bcbstxSource *handlers.BCBSTXSource) http.Handler {
	mux := http.NewServeMux()

	h := &handlers.Handler{DB: db, AetnaSource: aetnaSource, BCBSILSource: bcbsilSource, BCBSTXSource: bcbstxSource}

	mux.HandleFunc("GET /v1/reporting-plans/filters", h.GetReportingPlanFilters)
	mux.HandleFunc("GET /v1/reporting-plans", h.ListReportingPlans)
	mux.HandleFunc("GET /v1/indexes", h.ListIndexes)
	mux.HandleFunc("GET /v1/index-templates/payors", h.GetIndexTemplatePayors)
	mux.HandleFunc("GET /v1/index-templates", h.ListIndexTemplates)
	mux.HandleFunc("GET /v1/form-5500", h.ListForm5500)
	mux.HandleFunc("GET /v1/aetna-mrf/plans", h.GetAetnaMRFPlans)
	mux.HandleFunc("GET /v1/bcbsil-mrf/entries", h.GetBCBSILMRFEntries)
	mux.HandleFunc("GET /v1/bcbstx-mrf/entries", h.GetBCBSTXMRFEntries)

	return mux
}
