package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
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

type reportingPlanFilters struct {
	IngestorIDs     []uuid.UUID `json:"ingestor_ids"`
	PlanIDTypes     []string    `json:"plan_id_types"`
	PlanMarketTypes []string    `json:"plan_market_types"`
}

func (h *Handler) GetReportingPlanFilters(w http.ResponseWriter, r *http.Request) {
	filters, err := queryReportingPlanFilters(r.Context(), h.DB)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": map[string]string{
				"code":    "query_failed",
				"message": "failed to query reporting plan filters",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": filters,
	})
}

func (h *Handler) ListReportingPlans(w http.ResponseWriter, r *http.Request) {
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
	eins := parseStringFilter(r, "eins")

	normalizedEINs := make([]string, 0, len(eins)*2)
	for _, ein := range eins {
		plain, dashed := einVariants(ein)
		if plain == "" {
			continue
		}
		normalizedEINs = append(normalizedEINs, plain, dashed)
	}
	normalizedEINs = dedupeStrings(normalizedEINs)

	if len(ingestorIDs) == 0 || len(planIDTypes) == 0 || len(planMarketTypes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]string{
				"code":    "missing_filters",
				"message": "ingestor_ids, plan_id_types, and plan_market_types are required",
			},
		})
		return
	}

	plans, err := queryReportingPlans(r.Context(), h.DB, ingestorIDs, planIDTypes, planMarketTypes, normalizedEINs)
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

func queryReportingPlans(ctx context.Context, db *sql.DB, ingestorIDs []uuid.UUID, planIDTypes, planMarketTypes, eins []string) ([]reportingPlan, error) {
	if len(ingestorIDs) == 0 || len(planIDTypes) == 0 || len(planMarketTypes) == 0 {
		return nil, errors.New("all filters must be provided")
	}

	nextPlaceholder := 1
	ingestorPlaceholders, nextPlaceholder := placeholders(nextPlaceholder, len(ingestorIDs))
	planIDTypePlaceholders, nextPlaceholder := placeholders(nextPlaceholder, len(planIDTypes))
	planMarketTypePlaceholders, nextPlaceholder := placeholders(nextPlaceholder, len(planMarketTypes))

	einFilter := ""
	if len(eins) > 0 {
		einPlaceholders, next := placeholders(nextPlaceholder, len(eins))
		einFilter = fmt.Sprintf("\n\tand UPPER(rp.plan_id_type) = 'EIN'\n\tand rp.plan_id IN (%s)", einPlaceholders)
		nextPlaceholder = next
	}
	_ = nextPlaceholder

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
	%s
`, ingestorPlaceholders, planIDTypePlaceholders, planMarketTypePlaceholders, einFilter)

	args := make([]any, 0, len(ingestorIDs)+len(planIDTypes)+len(planMarketTypes)+len(eins))
	for _, id := range ingestorIDs {
		args = append(args, id)
	}
	for _, planIDType := range planIDTypes {
		args = append(args, planIDType)
	}
	for _, planMarketType := range planMarketTypes {
		args = append(args, planMarketType)
	}
	for _, ein := range eins {
		args = append(args, ein)
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

func queryReportingPlanFilters(ctx context.Context, db *sql.DB) (reportingPlanFilters, error) {
	filters := reportingPlanFilters{
		IngestorIDs:     make([]uuid.UUID, 0),
		PlanIDTypes:     make([]string, 0),
		PlanMarketTypes: make([]string, 0),
	}

	ingestorRows, err := db.QueryContext(ctx, `
select distinct ingestor_id
from indexes
where deleted_at is null
	and archived_at is null
order by ingestor_id
`)
	if err != nil {
		return filters, err
	}
	defer ingestorRows.Close()

	for ingestorRows.Next() {
		var id uuid.UUID
		if err := ingestorRows.Scan(&id); err != nil {
			return filters, err
		}
		filters.IngestorIDs = append(filters.IngestorIDs, id)
	}
	if err := ingestorRows.Err(); err != nil {
		return filters, err
	}

	planIDTypeRows, err := db.QueryContext(ctx, `
select distinct rp.plan_id_type
from reporting_plans rp
join indexes i on i.id = rp.index_id
where i.deleted_at is null
	and i.archived_at is null
	and rp.plan_id_type <> ''
order by rp.plan_id_type
`)
	if err != nil {
		return filters, err
	}
	defer planIDTypeRows.Close()

	for planIDTypeRows.Next() {
		var planIDType string
		if err := planIDTypeRows.Scan(&planIDType); err != nil {
			return filters, err
		}
		filters.PlanIDTypes = append(filters.PlanIDTypes, planIDType)
	}
	if err := planIDTypeRows.Err(); err != nil {
		return filters, err
	}

	planMarketTypeRows, err := db.QueryContext(ctx, `
select distinct rp.plan_market_type
from reporting_plans rp
join indexes i on i.id = rp.index_id
where i.deleted_at is null
	and i.archived_at is null
	and rp.plan_market_type <> ''
order by rp.plan_market_type
`)
	if err != nil {
		return filters, err
	}
	defer planMarketTypeRows.Close()

	for planMarketTypeRows.Next() {
		var planMarketType string
		if err := planMarketTypeRows.Scan(&planMarketType); err != nil {
			return filters, err
		}
		filters.PlanMarketTypes = append(filters.PlanMarketTypes, planMarketType)
	}
	if err := planMarketTypeRows.Err(); err != nil {
		return filters, err
	}

	return filters, nil
}

func dedupeStrings(values []string) []string {
	if len(values) <= 1 {
		return values
	}

	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
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
