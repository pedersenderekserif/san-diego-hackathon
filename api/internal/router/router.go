package router

import (
	"net/http"

	"github.com/pedersenderekserif/san-diego-hackathon/api/internal/handlers"
)

func New() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/reporting-plans", handlers.ListReportingPlans)

	return mux
}
