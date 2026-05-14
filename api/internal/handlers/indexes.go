package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/pedersenderekserif/san-diego-hackathon/api/internal/db"
)

type index struct {
	ID         uuid.UUID `json:"id"`
	IngestorID uuid.UUID `json:"ingestor_id"`
	DeletedAt  *string   `json:"deleted_at"`
	ArchivedAt *string   `json:"archived_at"`
}

func ListIndexes(w http.ResponseWriter, r *http.Request) {
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

	if len(ingestorIDs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]string{
				"code":    "missing_filters",
				"message": "ingestor_ids is required",
			},
		})
		return
	}

	includeDeleted := r.URL.Query().Get("include_deleted") == "true"
	includeArchived := r.URL.Query().Get("include_archived") == "true"

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

	indexes, err := queryIndexes(r.Context(), conn, ingestorIDs, includeDeleted, includeArchived)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": map[string]string{
				"code":    "query_failed",
				"message": "failed to query indexes",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": indexes,
		"meta": map[string]any{
			"count": len(indexes),
		},
	})
}

func queryIndexes(ctx context.Context, conn *sql.DB, ingestorIDs []uuid.UUID, includeDeleted, includeArchived bool) ([]index, error) {
	ingestorPlaceholders, nextPlaceholder := placeholders(1, len(ingestorIDs))

	conditions := []string{fmt.Sprintf("ingestor_id IN (%s)", ingestorPlaceholders)}
	if !includeDeleted {
		conditions = append(conditions, fmt.Sprintf("deleted_at IS NULL"))
	}
	if !includeArchived {
		conditions = append(conditions, fmt.Sprintf("archived_at IS NULL"))
	}
	_ = nextPlaceholder

	query := fmt.Sprintf(`
select id, ingestor_id, deleted_at, archived_at
from indexes
where %s
`, strings.Join(conditions, " and "))

	args := make([]any, len(ingestorIDs))
	for i, id := range ingestorIDs {
		args[i] = id
	}

	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]index, 0)
	for rows.Next() {
		var idx index
		if err := rows.Scan(&idx.ID, &idx.IngestorID, &idx.DeletedAt, &idx.ArchivedAt); err != nil {
			return nil, err
		}
		result = append(result, idx)
	}

	return result, rows.Err()
}
