# 🌊 San Diego Hackathon — Surf's Up on Self-Funded Employers 🏄

> *"The forecast is calling for 72°, sunny skies, and radical insights into employer health networks."*

---

## 🌴 What's the Vibe?

Serif Health's San Diego offsite hackathon — where we trade our laptops for a moment of sunshine (then immediately open them back up). This project digs into **self-funded employers** and the health networks they're plugged into, using real-world transparency data.

No wetsuits required. Just SQL, curiosity, and maybe a fish taco.

---

## 🐟 The Mission

Self-funded employers are a massive, underexplored segment of the US healthcare market. They pick their own networks — but which ones? Are they getting the best coverage bang for their buck? Do the networks they choose actually make sense for their workforce?

We're here to find out.

---

## 🏖️ Data Sources

| Source | What It Is |
|---|---|
| `reporting_plans` table | Insurance plans reporting under the Transparency in Coverage rule |
| `mrf-indexer` | Machine-Readable File index linking plans to their actual price data |
| **Form 5500 filings** | DOL filings from self-funded employers — our EIN goldmine 🥇 |
| `datasets/form_5500.zip` | Local snapshot of Form 5500 data for analysis |

---

## 🌊 How It Works

1. **Harvest EINs** from Form 5500 filings — these identify self-funded employers
2. **Cross-reference** with `reporting_plans` to find their declared networks
3. **Trace through** `mrf-indexer` to see what price data is actually published
4. **Analyze** — which networks are most common? Any surprises? Any gaps?

---

## 🛺 Getting Started

```bash
# Clone the repo (you're probably already here)
git clone https://github.com/pedersenderekserif/san-diego-hackathon
cd san-diego-hackathon
```

### 🐘 Start the Database

Spin up a local PostgreSQL instance and load all datasets in one shot:

```bash
docker compose up --build
```

This will:
1. Start a PostgreSQL 16 container on port `5432`
2. Create the schema (`indexes`, `index_templates`, `reporting_plans`, `form_5500`)
3. Load all datasets from `datasets/` automatically

**Connection details:**

| Setting  | Value       |
|----------|-------------|
| Host     | `localhost` |
| Port     | `5432`      |
| Database | `hackathon` |
| User     | `hackathon` |
| Password | `hackathon` |

Connect with psql:

```bash
psql "postgres://hackathon:hackathon@localhost:5432/hackathon"
```

### 🌐 Start the API

Set the DB env vars and run the Go API:

```bash
cd api
PG_HOST=localhost PG_USER=hackathon PG_PASSWORD=hackathon PG_DATABASE=hackathon make start
```

---

## 🌅 The Team

Built with ☀️ at Serif Health's San Diego offsite. Fueled by ocean air, good vibes, and a shared belief that healthcare pricing data should be a lot more legible than it currently is.

---

*"Hang ten on the data pipeline."* 🤙
