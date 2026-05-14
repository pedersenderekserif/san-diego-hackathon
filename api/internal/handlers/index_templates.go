package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type indexTemplate struct {
	IndexTemplateID uuid.UUID `json:"index_template_id"`
	PayorName       string    `json:"payor_name"`
	PayorID         uuid.UUID `json:"payor_id"`
}

type indexTemplatePayor struct {
	PayorID          uuid.UUID   `json:"payor_id"`
	PayorName        string      `json:"payor_name"`
	IndexTemplateIDs []uuid.UUID `json:"index_template_ids"`
}

func (h *Handler) GetIndexTemplatePayors(w http.ResponseWriter, r *http.Request) {
	payors, err := queryIndexTemplatePayors(r.Context(), h.DB)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": map[string]string{
				"code":    "query_failed",
				"message": "failed to query index template payors",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": payors,
		"meta": map[string]any{
			"count": len(payors),
		},
	})
}

func queryIndexTemplatePayors(ctx context.Context, conn *sql.DB) ([]indexTemplatePayor, error) {
	rows, err := conn.QueryContext(ctx, `
select payor_id, payor_name, index_template_id
from index_templates
order by payor_name, payor_id, index_template_id
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []indexTemplatePayor
	index := map[uuid.UUID]int{}

	for rows.Next() {
		var payorID, templateID uuid.UUID
		var payorName string
		if err := rows.Scan(&payorID, &payorName, &templateID); err != nil {
			return nil, err
		}

		if i, ok := index[payorID]; ok {
			result[i].IndexTemplateIDs = append(result[i].IndexTemplateIDs, templateID)
		} else {
			index[payorID] = len(result)
			result = append(result, indexTemplatePayor{
				PayorID:          payorID,
				PayorName:        payorName,
				IndexTemplateIDs: []uuid.UUID{templateID},
			})
		}
	}

	return result, rows.Err()
}

func (h *Handler) ListIndexTemplates(w http.ResponseWriter, r *http.Request) {
	payorIDs, err := parseUUIDFilter(r, "payor_ids")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]string{
				"code":    "invalid_request",
				"message": err.Error(),
			},
		})
		return
	}

	if len(payorIDs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]string{
				"code":    "missing_filters",
				"message": "payor_ids is required",
			},
		})
		return
	}

	templates, err := queryIndexTemplates(r.Context(), h.DB, payorIDs)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": map[string]string{
				"code":    "query_failed",
				"message": "failed to query index templates",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": templates,
		"meta": map[string]any{
			"count": len(templates),
		},
	})
}

func queryIndexTemplates(ctx context.Context, conn *sql.DB, payorIDs []uuid.UUID) ([]indexTemplate, error) {
	payorPlaceholders, _ := placeholders(1, len(payorIDs))

	query := fmt.Sprintf(`
select index_template_id, payor_name, payor_id
from index_templates
where payor_id IN (%s)
`, payorPlaceholders)

	args := make([]any, len(payorIDs))
	for i, id := range payorIDs {
		args[i] = id
	}

	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]indexTemplate, 0)
	for rows.Next() {
		var t indexTemplate
		if err := rows.Scan(&t.IndexTemplateID, &t.PayorName, &t.PayorID); err != nil {
			return nil, err
		}
		result = append(result, t)
	}

	return result, rows.Err()
}

func parseStringContainsFilter(r *http.Request, key string) string {
	return strings.TrimSpace(r.URL.Query().Get(key))
}
