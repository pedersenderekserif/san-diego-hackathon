package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
)

// normalizeEIN strips dashes and returns the 9-digit plain EIN if s is a valid
// EIN (either "XXXXXXXXX" or "XX-XXXXXXX"), otherwise returns "".
func normalizeEIN(s string) string {
	stripped := strings.ReplaceAll(strings.TrimSpace(s), "-", "")
	if len(stripped) != 9 {
		return ""
	}
	for _, c := range stripped {
		if c < '0' || c > '9' {
			return ""
		}
	}
	return stripped
}

// einVariants returns both the plain ("XXXXXXXXX") and dashed ("XX-XXXXXXX")
// forms of an EIN. Returns ("", "") if s does not look like a valid EIN.
func einVariants(s string) (plain, dashed string) {
	plain = normalizeEIN(s)
	if plain == "" {
		return "", ""
	}
	dashed = plain[:2] + "-" + plain[2:]
	return plain, dashed
}

type form5500 struct {
	AckID                 string `json:"ack_id"`
	FormPlanYearBeginDate string `json:"form_plan_year_begin_date"`
	FormTaxPrd            string `json:"form_tax_prd"`
	PlanName              string `json:"plan_name"`
	PlanEffDate           string `json:"plan_eff_date"`
	SponsorDfeName        string `json:"sponsor_dfe_name"`
	SponsDfeDbaName       string `json:"spons_dfe_dba_name"`
	SponsDfeEin           string `json:"spons_dfe_ein"`
	SponsDfeMailUsCity    string `json:"spons_dfe_mail_us_city"`
	SponsDfeMailUsState   string `json:"spons_dfe_mail_us_state"`
	SponsDfeMailUsZip     string `json:"spons_dfe_mail_us_zip"`
	TypePlanEntityCd      string `json:"type_plan_entity_cd"`
	TypeWelfareBnftCode   string `json:"type_welfare_bnft_code"`
	TotActRtdSepBenefCnt  string `json:"tot_act_rtd_sep_benef_cnt"`
	FilingStatus          string `json:"filing_status"`
	DateReceived          string `json:"date_received"`
}

func (h *Handler) ListForm5500(w http.ResponseWriter, r *http.Request) {
	eins := parseStringFilter(r, "eins")
	sponsorNames := parseStringFilter(r, "sponsor_names")
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	// Expand each EIN to both plain and dashed forms so either format matches.
	normalizedEINs := make([]string, 0, len(eins)*2)
	for _, ein := range eins {
		plain, dashed := einVariants(ein)
		if plain == "" {
			normalizedEINs = append(normalizedEINs, ein) // pass through non-EIN values unchanged
		} else {
			normalizedEINs = append(normalizedEINs, plain, dashed)
		}
	}
	eins = normalizedEINs

	if len(eins) == 0 && len(sponsorNames) == 0 && q == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]string{
				"code":    "missing_filters",
				"message": "at least one of eins, sponsor_names, or q is required",
			},
		})
		return
	}

	filings, err := queryForm5500(r.Context(), h.DB, eins, sponsorNames, q)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": map[string]string{
				"code":    "query_failed",
				"message": "failed to query form 5500 filings",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": filings,
		"meta": map[string]any{
			"count": len(filings),
		},
	})
}

func queryForm5500(ctx context.Context, conn *sql.DB, eins, sponsorNames []string, q string) ([]form5500, error) {
	args := make([]any, 0, len(eins)+len(sponsorNames)+2)
	conditions := make([]string, 0, 3)

	nextPlaceholder := 1

	if len(eins) > 0 {
		einPlaceholders, next := placeholders(nextPlaceholder, len(eins))
		conditions = append(conditions, fmt.Sprintf("spons_dfe_ein IN (%s)", einPlaceholders))
		nextPlaceholder = next
		for _, ein := range eins {
			args = append(args, ein)
		}
	}

	if len(sponsorNames) > 0 {
		namePlaceholders, next := placeholders(nextPlaceholder, len(sponsorNames))
		conditions = append(conditions, fmt.Sprintf("sponsor_dfe_name IN (%s)", namePlaceholders))
		nextPlaceholder = next
		for _, name := range sponsorNames {
			args = append(args, name)
		}
	}

	if q != "" {
		plain, dashed := einVariants(q)
		if plain != "" {
			// q looks like an EIN — match either stored format exactly.
			einPlaceholders, next := placeholders(nextPlaceholder, 2)
			conditions = append(conditions, fmt.Sprintf("spons_dfe_ein IN (%s)", einPlaceholders))
			args = append(args, plain, dashed)
			nextPlaceholder = next
		} else {
			like := "%" + q + "%"
			conditions = append(conditions, fmt.Sprintf(
				"(sponsor_dfe_name ILIKE $%d OR spons_dfe_ein ILIKE $%d)",
				nextPlaceholder, nextPlaceholder+1,
			))
			args = append(args, like, like)
		}
	}

	query := fmt.Sprintf(`
select
	ack_id
	, coalesce(form_plan_year_begin_date, '') as form_plan_year_begin_date
	, coalesce(form_tax_prd, '') as form_tax_prd
	, coalesce(plan_name, '') as plan_name
	, coalesce(plan_eff_date, '') as plan_eff_date
	, coalesce(sponsor_dfe_name, '') as sponsor_dfe_name
	, coalesce(spons_dfe_dba_name, '') as spons_dfe_dba_name
	, coalesce(spons_dfe_ein, '') as spons_dfe_ein
	, coalesce(spons_dfe_mail_us_city, '') as spons_dfe_mail_us_city
	, coalesce(spons_dfe_mail_us_state, '') as spons_dfe_mail_us_state
	, coalesce(spons_dfe_mail_us_zip, '') as spons_dfe_mail_us_zip
	, coalesce(type_plan_entity_cd, '') as type_plan_entity_cd
	, coalesce(type_welfare_bnft_code, '') as type_welfare_bnft_code
	, coalesce(tot_act_rtd_sep_benef_cnt, '') as tot_act_rtd_sep_benef_cnt
	, coalesce(filing_status, '') as filing_status
	, coalesce(date_received, '') as date_received
from form_5500
where %s
`, strings.Join(conditions, " or "))

	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]form5500, 0)
	for rows.Next() {
		var f form5500
		if err := rows.Scan(
			&f.AckID,
			&f.FormPlanYearBeginDate,
			&f.FormTaxPrd,
			&f.PlanName,
			&f.PlanEffDate,
			&f.SponsorDfeName,
			&f.SponsDfeDbaName,
			&f.SponsDfeEin,
			&f.SponsDfeMailUsCity,
			&f.SponsDfeMailUsState,
			&f.SponsDfeMailUsZip,
			&f.TypePlanEntityCd,
			&f.TypeWelfareBnftCode,
			&f.TotActRtdSepBenefCnt,
			&f.FilingStatus,
			&f.DateReceived,
		); err != nil {
			return nil, err
		}
		result = append(result, f)
	}

	return result, rows.Err()
}
