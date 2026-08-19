# Resilient Job Listing Ingestion Engine (Golang + Gin)

[![Live Demo](https://img.shields.io/badge/Render-Live%20Demo-success)](https://job-ingestion-service.onrender.com/api/jobs)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![Framework](https://img.shields.io/badge/Framework-Gin-008080)](https://gin-gonic.com/)

A high-performance, resilient data ingestion service built in **Golang (Gin framework)** designed to collect job listings from public feeds safely and reliably without getting blocked or causing silent pipeline failures.

---

## 🔗 Quick Links for Reviewers

- 🌐 **Live Deployed API**: [https://job-ingestion-service.onrender.com/api/jobs](https://job-ingestion-service.onrender.com/api/jobs)
- 📄 **1-Page Technical Decisions (`DECISIONS.md`)**: [View File](./DECISIONS.md)
- 📐 **Full Architectural Specification (`DESIGN.md`)**: [View File](./DESIGN.md)

---

## 🌟 Key Features

* **Humanized Request Pacing**: Uses random timing delay distributions (`300ms–900ms` jitter) between requests to break metronomic bot signatures.
* **User-Agent Pool & Header Alignment**: Emulates full modern desktop browser headers (`Chrome 125-127`, `Firefox 128`, `Safari 17.5`) including `Sec-Ch-Ua`, `Sec-Fetch-*`, and `Accept-Language`.
* **3-State Circuit Breaker**: Automatically transitions (`CLOSED` ➔ `OPEN` ➔ `HALF-OPEN`) when encountering rate limits (`429`) or blocks (`403`), protecting client IP reputation.
* **Fallback Cascade**: Primary Public Feed ➔ Secondary Public Feed ➔ Sandbox Mock Feed. Downstream consumers never experience 500 errors or zero-data crashes.
* **Defensive Schema Parsing**: Rejects zero-record malformed payloads to prevent overwriting existing valid cached job listings.

---

## 🚀 Live API Endpoints & Usage

### 1. Get Ingested Job Listings
Retrieve parsed and normalized job listings.

```bash
curl -X GET https://job-ingestion-service.onrender.com/api/jobs
```

**Filter by Keyword**:
```bash
curl -X GET "https://job-ingestion-service.onrender.com/api/jobs?q=engineer"
```

**Example Response**:
```json
{
  "count": 3,
  "jobs": [
    {
      "id": "sandbox-1",
      "title": "Senior Go Backend Engineer (Distributed Systems)",
      "company": "CloudScale Solutions",
      "location": "Remote / US-Only",
      "url": "https://example.com/jobs/sandbox-1",
      "description": "Designing resilient ingestion microservices with Go, Gin, and Docker.",
      "source": "Sandbox Mock Feed",
      "published_at": "2026-08-19T00:45:50Z",
      "fetched_at": "2026-08-19T02:45:50Z"
    }
  ]
}
```

### 2. Trigger On-Demand Ingestion
Triggers an immediate ingestion run across the fallback cascade.

```bash
curl -X POST https://job-ingestion-service.onrender.com/api/ingest
```

### 3. Check System & Circuit Breaker Health
```bash
curl -X GET https://job-ingestion-service.onrender.com/health
```

**Response**:
```json
{
  "active_source": "RemoteOK RSS",
  "circuit_breaker_status": "CLOSED",
  "service": "job-ingestion-engine",
  "status": "healthy"
}
```

### 4. Telemetry & Execution Metrics
```bash
curl -X GET https://job-ingestion-service.onrender.com/api/metrics
```

---

## 🛠️ Running Locally

### Option A: Using Go
```bash
# Clone repository
git clone https://github.com/theycallmesabb/acdyonassignment.git
cd acdyonassignment

# Run tests
go test -v ./...

# Start server
PORT=8080 go run ./cmd/server
```

### Option B: Using Docker
```bash
docker build -t job-service .
docker run -p 8080:8080 job-service
```

---

## 🏗️ Architecture Overview

```text
  +--------------------+       +----------------------+       +-------------------------+
  |  Client Request    | ----> |   Gin API Server     | ----> |  Resilient Client Engine |
  |  (Trigger/Cron)    |       |  (/api/ingest)       |       | (Jitter/UA Rotation)    |
  +--------------------+       +----------------------+       +-------------------------+
                                                                          |
                                                                          v
  +-------------------------------------------------------------------------------------+
  |                                SOURCE CASCADE ENGINE                                |
  |                                                                                     |
  |  +------------------------+    Fail/Block    +-----------------------------------+  |
  |  | Priority 1: Primary    | -------------->  | Priority 2: Secondary Public Feed |  |
  |  | Public Job Feed (RSS)  |                  | (WeWorkRemotely RSS)              |  |
  |  +------------------------+                  +-----------------------------------+  |
  |               |                                                |                    |
  |               | Failure / Circuit Breaker OPEN                 | Failure            |
  |               v                                                v                    |
  |  +-------------------------------------------------------------------------------+  |
  |  | Priority 3: Fallback Engine / Mock Sandbox Source                             |  |
  |  +-------------------------------------------------------------------------------+  |
  +-------------------------------------------------------------------------------------+
```
