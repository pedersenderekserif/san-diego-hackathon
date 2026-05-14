package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const aetnaMRFMetadataURL = "https://mrf.healthsparq.com/aetnacvs-egress.nophi.kyruushsq.com/prd/mrf/AETNACVS_I/ALICSI/latest_metadata.json"

type aetnaMRFReportingPlan struct {
	PlanName       string `json:"planName"`
	PlanIDType     string `json:"planIdType"`
	PlanID         string `json:"planId"`
	PlanMarketType string `json:"planMarketType"`
}

type aetnaMRFFile struct {
	ReportingEntityName string                  `json:"reportingEntityName"`
	ReportingPlans      []aetnaMRFReportingPlan `json:"reportingPlans"`
	LastUpdatedOn       string                  `json:"lastUpdatedOn"`
	FileSchema          string                  `json:"fileSchema"`
	FileName            string                  `json:"fileName"`
}

type aetnaMRFMetadata struct {
	Files []aetnaMRFFile `json:"files"`
}

// AetnaMRFPlan is a deduplicated plan entry exposed via the API.
type AetnaMRFPlan struct {
	PlanID         string `json:"plan_id"`
	PlanIDType     string `json:"plan_id_type"`
	PlanName       string `json:"plan_name"`
	PlanMarketType string `json:"plan_market_type"`
}

// AetnaSource holds the plans loaded from the Aetna MRF metadata at startup.
type AetnaSource struct {
	Plans []AetnaMRFPlan
}

// EINSet returns a set of all EINs (normalised, no dashes) in the source.
func (a *AetnaSource) EINSet() map[string]struct{} {
	if a == nil {
		return nil
	}
	set := make(map[string]struct{}, len(a.Plans))
	for _, p := range a.Plans {
		if p.PlanIDType == "ein" || p.PlanIDType == "EIN" {
			plain := normalizeEIN(p.PlanID)
			if plain != "" {
				set[plain] = struct{}{}
			}
		}
	}
	return set
}

// LoadAetnaSource fetches the Aetna MRF metadata URL and returns a deduplicated
// list of reporting plans. Non-fatal: logs and returns nil on error.
func LoadAetnaSource(ctx context.Context) *AetnaSource {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, aetnaMRFMetadataURL, nil)
	if err != nil {
		log.Printf("aetna_mrf: failed to build request: %v", err)
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("aetna_mrf: failed to fetch metadata: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("aetna_mrf: unexpected status %d", resp.StatusCode)
		return nil
	}

	var meta aetnaMRFMetadata
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		log.Printf("aetna_mrf: failed to decode metadata: %v", err)
		return nil
	}

	seen := make(map[string]struct{})
	plans := make([]AetnaMRFPlan, 0)
	for _, f := range meta.Files {
		for _, rp := range f.ReportingPlans {
			key := fmt.Sprintf("%s|%s", rp.PlanIDType, rp.PlanID)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			plans = append(plans, toAetnaMRFPlan(rp))
		}
	}

	log.Printf("aetna_mrf: loaded %d unique plans from MRF metadata", len(plans))
	return &AetnaSource{Plans: plans}
}

func toAetnaMRFPlan(rp aetnaMRFReportingPlan) AetnaMRFPlan {
	return AetnaMRFPlan{
		PlanID:         rp.PlanID,
		PlanIDType:     rp.PlanIDType,
		PlanName:       rp.PlanName,
		PlanMarketType: rp.PlanMarketType,
	}
}

// GetAetnaMRFPlans returns the in-memory Aetna MRF plans loaded at startup.
func (h *Handler) GetAetnaMRFPlans(w http.ResponseWriter, r *http.Request) {
	if h.AetnaSource == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": map[string]string{
				"code":    "source_unavailable",
				"message": "Aetna MRF source could not be loaded at startup",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": h.AetnaSource.Plans,
		"meta": map[string]any{
			"count": len(h.AetnaSource.Plans),
		},
	})
}
