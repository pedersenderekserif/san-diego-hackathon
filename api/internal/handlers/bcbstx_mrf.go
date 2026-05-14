package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

const bcbstxMRFFilelistURL = "https://www.bcbstx.com/content/dam/bcbs/mrf/si-filelist.json"

// bcbstxPayorID is the well-known payor_id (== ingestor_id) for BCBSTX.
// When this payor is requested, EIN data is sourced from the in-memory BCBSTX
// filelist instead of the reporting_plans/indexes DB subquery.
const bcbstxPayorID = "c467dfff-cb79-4f3b-b407-6d3f3e9f9bb6"

// BCBSTXEntry is a deduplicated filelist entry exposed via the API.
type BCBSTXEntry struct {
	EIN            string `json:"ein"`
	Name           string `json:"name"`
	State          string `json:"state"`
	URL            string `json:"url"`
	LastUpdateDate string `json:"last_update_date"`
}

// BCBSTXSource holds the entries loaded from the BCBSTX MRF filelist at startup.
type BCBSTXSource struct {
	Entries []BCBSTXEntry
}

// EINSet returns a set of all normalised (no-dash) EINs in the source.
func (b *BCBSTXSource) EINSet() map[string]struct{} {
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

// LoadBCBSTXSource fetches the BCBSTX MRF filelist URL and returns a deduplicated
// list of entries keyed by EIN. Non-fatal: logs and returns nil on error.
func LoadBCBSTXSource(ctx context.Context) *BCBSTXSource {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, bcbstxMRFFilelistURL, nil)
	if err != nil {
		log.Printf("bcbstx_mrf: failed to build request: %v", err)
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("bcbstx_mrf: failed to fetch filelist: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("bcbstx_mrf: unexpected status %d", resp.StatusCode)
		return nil
	}

	// Reuse the same JSON structure as BCBSIL.
	var raw []bcbsilFileEntry
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		log.Printf("bcbstx_mrf: failed to decode filelist: %v", err)
		return nil
	}

	seen := make(map[string]struct{})
	entries := make([]BCBSTXEntry, 0, len(raw))
	for _, e := range raw {
		plain := normalizeEIN(e.EIN)
		if plain == "" {
			continue
		}
		if _, ok := seen[plain]; ok {
			continue
		}
		seen[plain] = struct{}{}
		entries = append(entries, BCBSTXEntry{
			EIN:            e.EIN,
			Name:           e.Name,
			State:          e.State,
			URL:            e.URL,
			LastUpdateDate: e.LastUpdateDate,
		})
	}

	log.Printf("bcbstx_mrf: loaded %d unique entries from filelist", len(entries))
	return &BCBSTXSource{Entries: entries}
}

// GetBCBSTXMRFEntries returns the in-memory BCBSTX MRF entries loaded at startup.
func (h *Handler) GetBCBSTXMRFEntries(w http.ResponseWriter, r *http.Request) {
	if h.BCBSTXSource == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": map[string]string{
				"code":    "source_unavailable",
				"message": "BCBSTX MRF source could not be loaded at startup",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": h.BCBSTXSource.Entries,
		"meta": map[string]any{
			"count": len(h.BCBSTXSource.Entries),
		},
	})
}
