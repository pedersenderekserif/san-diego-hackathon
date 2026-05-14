package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

const bcbsilMRFFilelistURL = "https://www.bcbsil.com/content/dam/bcbs/mrf/si-filelist.json"

// bcbsilPayorID is the well-known payor_id (== ingestor_id) for BCBSIL.
// When this payor is requested, EIN data is sourced from the in-memory BCBSIL
// filelist instead of the reporting_plans/indexes DB subquery.
const bcbsilPayorID = "24b8e840-2a4e-46ee-b52c-82b21bb60427"

type bcbsilFileEntry struct {
	LastUpdateDate string `json:"last_update_date"`
	State          string `json:"state"`
	URL            string `json:"url"`
	Name           string `json:"name"`
	EIN            string `json:"ein"`
}

// BCBSILEntry is a deduplicated filelist entry exposed via the API.
type BCBSILEntry struct {
	EIN            string `json:"ein"`
	Name           string `json:"name"`
	State          string `json:"state"`
	URL            string `json:"url"`
	LastUpdateDate string `json:"last_update_date"`
}

// BCBSILSource holds the entries loaded from the BCBSIL MRF filelist at startup.
type BCBSILSource struct {
	Entries []BCBSILEntry
}

// EINSet returns a set of all normalised (no-dash) EINs in the source.
func (b *BCBSILSource) EINSet() map[string]struct{} {
	if b == nil {
		return nil
	}
	set := make(map[string]struct{}, len(b.Entries))
	for _, e := range b.Entries {
		plain := normalizeEIN(e.EIN)
		if plain != "" {
			set[plain] = struct{}{}
		}
	}
	return set
}

// LoadBCBSILSource fetches the BCBSIL MRF filelist URL and returns a deduplicated
// list of entries keyed by EIN. Non-fatal: logs and returns nil on error.
func LoadBCBSILSource(ctx context.Context) *BCBSILSource {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, bcbsilMRFFilelistURL, nil)
	if err != nil {
		log.Printf("bcbsil_mrf: failed to build request: %v", err)
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("bcbsil_mrf: failed to fetch filelist: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("bcbsil_mrf: unexpected status %d", resp.StatusCode)
		return nil
	}

	var raw []bcbsilFileEntry
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		log.Printf("bcbsil_mrf: failed to decode filelist: %v", err)
		return nil
	}

	seen := make(map[string]struct{})
	entries := make([]BCBSILEntry, 0, len(raw))
	for _, e := range raw {
		plain := normalizeEIN(e.EIN)
		if plain == "" {
			continue
		}
		if _, ok := seen[plain]; ok {
			continue
		}
		seen[plain] = struct{}{}
		entries = append(entries, BCBSILEntry{
			EIN:            e.EIN,
			Name:           e.Name,
			State:          e.State,
			URL:            e.URL,
			LastUpdateDate: e.LastUpdateDate,
		})
	}

	log.Printf("bcbsil_mrf: loaded %d unique entries from filelist", len(entries))
	return &BCBSILSource{Entries: entries}
}

// GetBCBSILMRFEntries returns the in-memory BCBSIL MRF entries loaded at startup.
func (h *Handler) GetBCBSILMRFEntries(w http.ResponseWriter, r *http.Request) {
	if h.BCBSILSource == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": map[string]string{
				"code":    "source_unavailable",
				"message": "BCBSIL MRF source could not be loaded at startup",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": h.BCBSILSource.Entries,
		"meta": map[string]any{
			"count": len(h.BCBSILSource.Entries),
		},
	})
}
