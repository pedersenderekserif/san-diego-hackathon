package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
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

// aetnaPayorID is the well-known payor_id (== ingestor_id) for Aetna. When this
// payor is requested, form 5500 results are scoped using the in-memory Aetna MRF
// EIN set rather than the reporting_plans/indexes DB subquery.
const aetnaPayorID = "9b29c0e6-21b4-41b0-bd13-a7d8a4342d4c"

func (h *Handler) ListForm5500(w http.ResponseWriter, r *http.Request) {
	eins := parseStringFilter(r, "eins")
	sponsorNames := parseStringFilter(r, "sponsor_names")
	var payorIDs []string
	isAetna := false
	isBCBSIL := false
	isBCBSTX := false
	if rawPayorID := strings.TrimSpace(r.URL.Query().Get("payor_id")); rawPayorID != "" {
		parsed, err := uuid.Parse(rawPayorID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"error": map[string]string{
					"code":    "invalid_filters",
					"message": "payor_id must be a valid UUID",
				},
			})
			return
		}
		switch parsed.String() {
		case aetnaPayorID:
			isAetna = true
		case bcbsilPayorID:
			isBCBSIL = true
		case bcbstxPayorID:
			isBCBSTX = true
		default:
			payorIDs = []string{parsed.String()}
		}
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	fundingGenAsset := r.URL.Query().Get("funding_gen_asset_ind") == "1"
	// Normalize EINs to plain 9-digit form only; SQL normalizes the column with replace().
	normalizedEINs := make([]string, 0, len(eins))
	for _, ein := range eins {
		plain := normalizeEIN(ein)
		if plain == "" {
			normalizedEINs = append(normalizedEINs, ein) // pass through non-EIN values unchanged
		} else {
			normalizedEINs = append(normalizedEINs, plain)
		}
	}
	eins = normalizedEINs

	if len(eins) == 0 && len(sponsorNames) == 0 && q == "" && len(payorIDs) == 0 && !isAetna && !isBCBSIL && !isBCBSTX {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]string{
				"code":    "missing_filters",
				"message": "at least one of eins, sponsor_names, q, or payor_id is required",
			},
		})
		return
	}

	var filings []form5500
	var err error
	if isAetna {
		if h.AetnaSource == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"error": map[string]string{
					"code":    "source_unavailable",
					"message": "Aetna MRF source could not be loaded at startup",
				},
			})
			return
		}
		filings, err = queryForm5500Aetna(r.Context(), h.DB, eins, sponsorNames, q, h.AetnaSource.EINSet(), fundingGenAsset)
	} else if isBCBSIL {
		if h.BCBSILSource == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"error": map[string]string{
					"code":    "source_unavailable",
					"message": "BCBSIL MRF source could not be loaded at startup",
				},
			})
			return
		}
		filings, err = queryForm5500BCBSIL(r.Context(), h.DB, eins, sponsorNames, q, h.BCBSILSource.EINSet(), fundingGenAsset)
	} else if isBCBSTX {
		if h.BCBSTXSource == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"error": map[string]string{
					"code":    "source_unavailable",
					"message": "BCBSTX MRF source could not be loaded at startup",
				},
			})
			return
		}
		filings, err = queryForm5500BCBSTX(r.Context(), h.DB, eins, sponsorNames, q, h.BCBSTXSource.EINSet(), fundingGenAsset)
	} else {
		filings, err = queryForm5500(r.Context(), h.DB, eins, sponsorNames, q, payorIDs, fundingGenAsset)
	}
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

func queryForm5500(ctx context.Context, conn *sql.DB, eins, sponsorNames []string, q string, payorIDs []string, fundingGenAsset bool) ([]form5500, error) {
	args := make([]any, 0, len(eins)+len(sponsorNames)+2)
	searchConditions := make([]string, 0, 3)
	scopeConditions := make([]string, 0, 1)

	nextPlaceholder := 1

	if len(eins) > 0 {
		einPlaceholders, next := placeholders(nextPlaceholder, len(eins))
		searchConditions = append(searchConditions, fmt.Sprintf("replace(spons_dfe_ein, '-', '') IN (%s)", einPlaceholders))
		nextPlaceholder = next
		for _, ein := range eins {
			args = append(args, ein)
		}
	}

	if len(sponsorNames) > 0 {
		namePlaceholders, next := placeholders(nextPlaceholder, len(sponsorNames))
		searchConditions = append(searchConditions, fmt.Sprintf("sponsor_dfe_name IN (%s)", namePlaceholders))
		nextPlaceholder = next
		for _, name := range sponsorNames {
			args = append(args, name)
		}
	}

	if q != "" {
		plain, _ := einVariants(q)
		if plain != "" {
			// q looks like an EIN — normalize both sides so either stored format matches.
			searchConditions = append(searchConditions, fmt.Sprintf("replace(spons_dfe_ein, '-', '') = $%d", nextPlaceholder))
			args = append(args, plain)
			nextPlaceholder++
		} else {
			like := "%" + q + "%"
			searchConditions = append(searchConditions, fmt.Sprintf(
				"(sponsor_dfe_name ILIKE $%d OR spons_dfe_ein ILIKE $%d)",
				nextPlaceholder, nextPlaceholder+1,
			))
			args = append(args, like, like)
			nextPlaceholder += 2
		}
	}

	if len(payorIDs) > 0 {
		payorPlaceholderParts := make([]string, len(payorIDs))
		for i := range payorIDs {
			payorPlaceholderParts[i] = fmt.Sprintf("$%d::uuid", nextPlaceholder+i)
		}
		payorPlaceholders := strings.Join(payorPlaceholderParts, ",")
		nextPlaceholder += len(payorIDs)
		scopeConditions = append(scopeConditions, fmt.Sprintf(`
			spons_dfe_ein <> ''
			and exists (
				select 1
				from reporting_plans rp
				join indexes i on i.id = rp.index_id
				where i.deleted_at is null
					and i.archived_at is null
					and i.ingestor_id in (%s)
					and UPPER(rp.plan_id_type) = 'EIN'
					and replace(rp.plan_id, '-', '') = replace(form_5500.spons_dfe_ein, '-', '')
			)
		`, payorPlaceholders))
		for _, payorID := range payorIDs {
			args = append(args, payorID)
		}
	}

	whereParts := make([]string, 0, 3)
	if fundingGenAsset {
		whereParts = append(whereParts, "funding_gen_asset_ind = '1'")
	}
	if len(searchConditions) > 0 {
		whereParts = append(whereParts, "("+strings.Join(searchConditions, " or ")+")")
	}
	if len(scopeConditions) > 0 {
		whereParts = append(whereParts, strings.Join(scopeConditions, " and "))
	}

	query := fmt.Sprintf(`
select distinct on (replace(spons_dfe_ein, '-', ''))
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
order by replace(spons_dfe_ein, '-', ''), date_received desc nulls last
`, strings.Join(whereParts, " and "))

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

// queryForm5500Aetna queries form_5500 records scoped to the EINs present in the
// Aetna MRF in-memory source, instead of going through the reporting_plans/indexes
// DB subquery used for other payors.
func queryForm5500Aetna(ctx context.Context, conn *sql.DB, eins, sponsorNames []string, q string, aetnaEINs map[string]struct{}, fundingGenAsset bool) ([]form5500, error) {
	if len(aetnaEINs) == 0 {
		return []form5500{}, nil
	}

	args := make([]any, 0)
	searchConditions := make([]string, 0, 3)
	nextPlaceholder := 1

	// Scope: restrict to EINs present in the Aetna MRF source using normalized comparison.
	mrfEINSlice := make([]string, 0, len(aetnaEINs))
	for plain := range aetnaEINs {
		mrfEINSlice = append(mrfEINSlice, plain)
	}
	mrfPlaceholders, next := placeholders(nextPlaceholder, len(mrfEINSlice))
	nextPlaceholder = next
	scopeCondition := fmt.Sprintf("replace(spons_dfe_ein, '-', '') IN (%s)", mrfPlaceholders)
	for _, e := range mrfEINSlice {
		args = append(args, e)
	}

	if len(eins) > 0 {
		einPlaceholders, next := placeholders(nextPlaceholder, len(eins))
		searchConditions = append(searchConditions, fmt.Sprintf("replace(spons_dfe_ein, '-', '') IN (%s)", einPlaceholders))
		nextPlaceholder = next
		for _, ein := range eins {
			args = append(args, ein)
		}
	}

	if len(sponsorNames) > 0 {
		namePlaceholders, next := placeholders(nextPlaceholder, len(sponsorNames))
		searchConditions = append(searchConditions, fmt.Sprintf("sponsor_dfe_name IN (%s)", namePlaceholders))
		nextPlaceholder = next
		for _, name := range sponsorNames {
			args = append(args, name)
		}
	}

	if q != "" {
		plain, _ := einVariants(q)
		if plain != "" {
			searchConditions = append(searchConditions, fmt.Sprintf("replace(spons_dfe_ein, '-', '') = $%d", nextPlaceholder))
			args = append(args, plain)
			nextPlaceholder++
		} else {
			like := "%" + q + "%"
			searchConditions = append(searchConditions, fmt.Sprintf(
				"(sponsor_dfe_name ILIKE $%d OR spons_dfe_ein ILIKE $%d)",
				nextPlaceholder, nextPlaceholder+1,
			))
			args = append(args, like, like)
			nextPlaceholder += 2
		}
	}

	_ = nextPlaceholder

	whereParts := []string{scopeCondition}
	if fundingGenAsset {
		whereParts = append(whereParts, "funding_gen_asset_ind = '1'")
	}
	if len(searchConditions) > 0 {
		whereParts = append(whereParts, "("+strings.Join(searchConditions, " or ")+")")
	}

	query := fmt.Sprintf(`
select distinct on (replace(spons_dfe_ein, '-', ''))
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
order by replace(spons_dfe_ein, '-', ''), date_received desc nulls last
`, strings.Join(whereParts, " and "))

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

// queryForm5500BCBSIL queries form_5500 records scoped to the EINs present in
// the BCBSIL MRF in-memory source, instead of going through the
// reporting_plans/indexes DB subquery used for other payors.
func queryForm5500BCBSIL(ctx context.Context, conn *sql.DB, eins, sponsorNames []string, q string, bcbsilEINs map[string]struct{}, fundingGenAsset bool) ([]form5500, error) {
	if len(bcbsilEINs) == 0 {
		return []form5500{}, nil
	}

	args := make([]any, 0)
	searchConditions := make([]string, 0, 3)
	nextPlaceholder := 1

	// Scope: restrict to EINs present in the BCBSIL MRF filelist using normalized comparison.
	mrfEINSlice := make([]string, 0, len(bcbsilEINs))
	for plain := range bcbsilEINs {
		mrfEINSlice = append(mrfEINSlice, plain)
	}
	mrfPlaceholders, next := placeholders(nextPlaceholder, len(mrfEINSlice))
	nextPlaceholder = next
	scopeCondition := fmt.Sprintf("replace(spons_dfe_ein, '-', '') IN (%s)", mrfPlaceholders)
	for _, e := range mrfEINSlice {
		args = append(args, e)
	}

	if len(eins) > 0 {
		einPlaceholders, next := placeholders(nextPlaceholder, len(eins))
		searchConditions = append(searchConditions, fmt.Sprintf("replace(spons_dfe_ein, '-', '') IN (%s)", einPlaceholders))
		nextPlaceholder = next
		for _, ein := range eins {
			args = append(args, ein)
		}
	}

	if len(sponsorNames) > 0 {
		namePlaceholders, next := placeholders(nextPlaceholder, len(sponsorNames))
		searchConditions = append(searchConditions, fmt.Sprintf("sponsor_dfe_name IN (%s)", namePlaceholders))
		nextPlaceholder = next
		for _, name := range sponsorNames {
			args = append(args, name)
		}
	}

	if q != "" {
		plain, _ := einVariants(q)
		if plain != "" {
			searchConditions = append(searchConditions, fmt.Sprintf("replace(spons_dfe_ein, '-', '') = $%d", nextPlaceholder))
			args = append(args, plain)
			nextPlaceholder++
		} else {
			like := "%" + q + "%"
			searchConditions = append(searchConditions, fmt.Sprintf(
				"(sponsor_dfe_name ILIKE $%d OR spons_dfe_ein ILIKE $%d)",
				nextPlaceholder, nextPlaceholder+1,
			))
			args = append(args, like, like)
			nextPlaceholder += 2
		}
	}

	_ = nextPlaceholder

	whereParts := []string{scopeCondition}
	if fundingGenAsset {
		whereParts = append(whereParts, "funding_gen_asset_ind = '1'")
	}
	if len(searchConditions) > 0 {
		whereParts = append(whereParts, "("+strings.Join(searchConditions, " or ")+")")
	}

	query := fmt.Sprintf(`
select distinct on (replace(spons_dfe_ein, '-', ''))
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
order by replace(spons_dfe_ein, '-', ''), date_received desc nulls last
`, strings.Join(whereParts, " and "))

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

// queryForm5500BCBSTX queries form_5500 records scoped to the EINs present in
// the BCBSTX MRF in-memory source, instead of going through the
// reporting_plans/indexes DB subquery used for other payors.
func queryForm5500BCBSTX(ctx context.Context, conn *sql.DB, eins, sponsorNames []string, q string, bcbstxEINs map[string]struct{}, fundingGenAsset bool) ([]form5500, error) {
	if len(bcbstxEINs) == 0 {
		return []form5500{}, nil
	}

	args := make([]any, 0)
	searchConditions := make([]string, 0, 3)
	nextPlaceholder := 1

	// Scope: restrict to EINs present in the BCBSTX MRF filelist using normalized comparison.
	mrfEINSlice := make([]string, 0, len(bcbstxEINs))
	for plain := range bcbstxEINs {
		mrfEINSlice = append(mrfEINSlice, plain)
	}
	mrfPlaceholders, next := placeholders(nextPlaceholder, len(mrfEINSlice))
	nextPlaceholder = next
	scopeCondition := fmt.Sprintf("replace(spons_dfe_ein, '-', '') IN (%s)", mrfPlaceholders)
	for _, e := range mrfEINSlice {
		args = append(args, e)
	}

	if len(eins) > 0 {
		einPlaceholders, next := placeholders(nextPlaceholder, len(eins))
		searchConditions = append(searchConditions, fmt.Sprintf("replace(spons_dfe_ein, '-', '') IN (%s)", einPlaceholders))
		nextPlaceholder = next
		for _, ein := range eins {
			args = append(args, ein)
		}
	}

	if len(sponsorNames) > 0 {
		namePlaceholders, next := placeholders(nextPlaceholder, len(sponsorNames))
		searchConditions = append(searchConditions, fmt.Sprintf("sponsor_dfe_name IN (%s)", namePlaceholders))
		nextPlaceholder = next
		for _, name := range sponsorNames {
			args = append(args, name)
		}
	}

	if q != "" {
		plain, _ := einVariants(q)
		if plain != "" {
			searchConditions = append(searchConditions, fmt.Sprintf("replace(spons_dfe_ein, '-', '') = $%d", nextPlaceholder))
			args = append(args, plain)
			nextPlaceholder++
		} else {
			like := "%" + q + "%"
			searchConditions = append(searchConditions, fmt.Sprintf(
				"(sponsor_dfe_name ILIKE $%d OR spons_dfe_ein ILIKE $%d)",
				nextPlaceholder, nextPlaceholder+1,
			))
			args = append(args, like, like)
			nextPlaceholder += 2
		}
	}

	_ = nextPlaceholder

	whereParts := []string{scopeCondition}
	if fundingGenAsset {
		whereParts = append(whereParts, "funding_gen_asset_ind = '1'")
	}
	if len(searchConditions) > 0 {
		whereParts = append(whereParts, "("+strings.Join(searchConditions, " or ")+")")
	}

	query := fmt.Sprintf(`
select distinct on (replace(spons_dfe_ein, '-', ''))
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
order by replace(spons_dfe_ein, '-', ''), date_received desc nulls last
`, strings.Join(whereParts, " and "))

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
