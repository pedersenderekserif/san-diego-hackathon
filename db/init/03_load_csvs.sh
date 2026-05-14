#!/usr/bin/env bash
set -euo pipefail

echo "==> Loading datasets into $POSTGRES_DB..."

# index_templates (plain CSV, no unzip needed)
echo "  Loading index_templates..."
psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" <<'SQL'
CREATE TEMP TABLE tmp_index_templates (LIKE index_templates);
\COPY tmp_index_templates (index_template_id, payor_name, payor_id) FROM '/datasets/index_templates.csv' WITH CSV HEADER
INSERT INTO index_templates SELECT * FROM tmp_index_templates ON CONFLICT DO NOTHING;
SQL

# reporting_plans (zipped CSV)
echo "  Extracting reporting_plans.csv..."
unzip -p /datasets/reporting_plans.csv.zip reporting_plans.csv > /tmp/reporting_plans.csv

echo "  Loading reporting_plans..."
psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
    -c "\COPY reporting_plans FROM '/tmp/reporting_plans.csv' WITH CSV DELIMITER E'\t' QUOTE E'\b' ESCAPE E'\\\\' HEADER"

rm -f /tmp/reporting_plans.csv

# form_5500 (zipped CSV)
echo "  Extracting form_5500 CSV..."
unzip -p /datasets/form_5500.zip form_5500/2025/f_5500_2025_latest.csv > /tmp/form_5500.csv

echo "  Loading form_5500..."
psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
    -c "\COPY form_5500 FROM '/tmp/form_5500.csv' WITH CSV HEADER"

rm -f /tmp/form_5500.csv

echo "==> Dataset loading complete."
