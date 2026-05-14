package router

import (
	"net/http"

	"github.com/pedersenderekserif/san-diego-hackathon/api/internal/handlers"
)

func New() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/reporting-plans", handlers.ListReportingPlans)
	mux.HandleFunc("GET /v1/indexes", handlers.ListIndexes)
	mux.HandleFunc("GET /v1/index-templates", handlers.ListIndexTemplates)
	mux.HandleFunc("GET /v1/form-5500", handlers.ListForm5500)

	return mux
}
