package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/pedersenderekserif/san-diego-hackathon/api/internal/db"
)

type reportingPlan struct {
	ID              uuid.UUID `json:"id"`
	IndexID         uuid.UUID `json:"index_id"`
	RunID           uuid.UUID `json:"run_id"`
	FileID          uuid.UUID `json:"file_id"`
	PlanID          string    `json:"plan_id"`
	PlanName        string    `json:"plan_name"`
	PlanIDType      string    `json:"plan_id_type"`
	PlanMarketType  string    `json:"plan_market_type"`
	CreatedAt       string    `json:"created_at"`
	IssuerName      string    `json:"issuer_name"`
	PlanSponsorName string    `json:"plan_sponsor_name"`
}

func ListReportingPlans(w http.ResponseWriter, r *http.Request) {
	ingestorIDs, err := parseUUIDFilter(r, "ingestor_ids")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]string{
				"code":    "invalid_request",
				"message": err.Error(),
			},
		})
		return
	}

	planIDTypes := parseStringFilter(r, "plan_id_types")
	planMarketTypes := parseStringFilter(r, "plan_market_types")

	if len(ingestorIDs) == 0 || len(planIDTypes) == 0 || len(planMarketTypes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]string{
				"code":    "missing_filters",
				"message": "ingestor_ids, plan_id_types, and plan_market_types are required",
			},
		})
		return
	}

	conn, err := db.NewPostgresFromEnv(r.Context())
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": map[string]string{
				"code":    "db_not_configured",
				"message": "set PG_HOST, PG_USER, and PG_PASSWORD to enable this endpoint",
			},
		})
		return
	}
	defer conn.Close()

	plans, err := queryReportingPlans(r.Context(), conn, ingestorIDs, planIDTypes, planMarketTypes)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": map[string]string{
				"code":    "query_failed",
				"message": "failed to query reporting plans",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": plans,
		"meta": map[string]any{
			"count": len(plans),
		},
	})
}

func queryReportingPlans(ctx context.Context, db *sql.DB, ingestorIDs []uuid.UUID, planIDTypes, planMarketTypes []string) ([]reportingPlan, error) {
	if len(ingestorIDs) == 0 || len(planIDTypes) == 0 || len(planMarketTypes) == 0 {
		return nil, errors.New("all filters must be provided")
	}

	nextPlaceholder := 1
	ingestorPlaceholders, nextPlaceholder := placeholders(nextPlaceholder, len(ingestorIDs))
	planIDTypePlaceholders, nextPlaceholder := placeholders(nextPlaceholder, len(planIDTypes))
	planMarketTypePlaceholders, _ := placeholders(nextPlaceholder, len(planMarketTypes))

	query := fmt.Sprintf(`
select
	rp.id
	, rp.index_id
	, rp.run_id
	, rp.file_id
	, rp.plan_id
	, rp.plan_name
	, rp.plan_id_type
	, rp.plan_market_type
	, rp.created_at
	, rp.issuer_name
	, rp.plan_sponsor_name
from reporting_plans rp
where rp.index_id IN (
	select
		id
	from indexes
	where ingestor_id IN (%s)
		and deleted_at is null
		and archived_at is null
	)
	and rp.plan_id_type IN (%s)
	and rp.plan_market_type IN (%s)
`, ingestorPlaceholders, planIDTypePlaceholders, planMarketTypePlaceholders)

	args := make([]any, 0, len(ingestorIDs)+len(planIDTypes)+len(planMarketTypes))
	for _, id := range ingestorIDs {
		args = append(args, id)
	}
	for _, planIDType := range planIDTypes {
		args = append(args, planIDType)
	}
	for _, planMarketType := range planMarketTypes {
		args = append(args, planMarketType)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plans := make([]reportingPlan, 0)
	for rows.Next() {
		var plan reportingPlan
		if err := rows.Scan(
			&plan.ID,
			&plan.IndexID,
			&plan.RunID,
			&plan.FileID,
			&plan.PlanID,
			&plan.PlanName,
			&plan.PlanIDType,
			&plan.PlanMarketType,
			&plan.CreatedAt,
			&plan.IssuerName,
			&plan.PlanSponsorName,
		); err != nil {
			return nil, err
		}

		plans = append(plans, plan)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}

func parseUUIDFilter(r *http.Request, key string) ([]uuid.UUID, error) {
	values := parseStringFilter(r, key)
	ids := make([]uuid.UUID, 0, len(values))

	for _, value := range values {
		id, err := uuid.Parse(value)
		if err != nil {
			return nil, fmt.Errorf("%s must contain UUID values", key)
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func parseStringFilter(r *http.Request, key string) []string {
	rawValues, ok := r.URL.Query()[key]
	if !ok {
		return nil
	}

	values := make([]string, 0, len(rawValues))
	for _, raw := range rawValues {
		for _, part := range strings.Split(raw, ",") {
			value := strings.TrimSpace(part)
			if value == "" {
				continue
			}

			values = append(values, value)
		}
	}

	return values
}

func placeholders(start, n int) (string, int) {
	if n <= 0 {
		return "", start
	}

	parts := make([]string, n)
	for i := range parts {
		parts[i] = fmt.Sprintf("$%d", start+i)
	}

	return strings.Join(parts, ","), start + n
}
