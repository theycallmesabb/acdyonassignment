# Job Ingestion Engine: Architectural Design & Resilience Specification

## Executive Overview
This document outlines the architectural blueprint, detection surface analysis, ingestion strategies, resilience mechanisms, and ethical boundaries for automated job listing extraction across high-protection platforms (LinkedIn, Indeed, Naukri, Wellfound).

---

```
                                  INGESTION PIPELINE ARCHITECTURE
                                  
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
                                          |
                                          v
                              +------------------------+
                              | Normalized Job Storage |
                              | (/api/jobs Data API)   |
                              +------------------------+
```

---

## 1. Detection Surface: What Unmasks Automated Clients

High-volume platforms employ anti-bot systems (Cloudflare Bot Management, Datadome, Kasada, Akamai Bot Manager, PerimeterX) operating across network, transport, application, and client execution layers.

| Layer | Detection Surface | Mechanism / Indicator | Mitigated In Service Architecture |
| :--- | :--- | :--- | :--- |
| **TLS/Transport** | **JA3 / JA4 Fingerprinting** | Go `net/http` TLS Client Hello has a fixed cipher suite cipher order and extension list different from standard Chrome/Firefox browsers. | Abstracted behind proxy wrappers / TLS fingerprint matching options. |
| **HTTP/2** | **SETTINGS & WINDOW_UPDATE Frames** | Browser engines send HTTP/2 SETTINGS frames with unique pseudo-header ordering (`:method`, `:authority`, `:scheme`, `:path`). | Standard header ordering enforced via humanized request builder (`internal/ingestor/headers.go`). |
| **HTTP Application** | **Header Anomalies & Missing Headers** | Automated tools omit standard browser headers like `Sec-Ch-Ua`, `Sec-Fetch-Dest`, `Sec-Fetch-Mode`, `Sec-Fetch-Site`, `Accept-Language`, or `Accept-Encoding`. | Strict emulation of Chrome/macOS browser headers in `BuildHumanizedHeaders()`. |
| **Execution Environment** | **Headless Browser CDP Artifacts** | `navigator.webdriver = true`, missing `window.chrome`, inconsistent Canvas/WebGL rendering hashes, evaluation of Chrome DevTools Protocol (CDP) properties. | Preferred raw HTTP ingestion pattern over headless automation to eliminate DOM/CDP attack surface. |
| **Temporal / Behavioral** | **Metronomic Timing Patterns** | Sub-second request spacing with zero variance (e.g., exact 1000ms intervals) triggers statistical anomaly detection algorithms. | Random delay distribution (300ms–900ms jitter) built into `HumanizedDelay()`. |
| **IP Layer** | **IP Reputation & ASN Classification** | Datacenter ASNs (AWS, GCP, DigitalOcean, Hetzner) are pre-flagged. High request concurrency from single IP triggers automatic rate limits. | Proxy pool rotation architecture ready (residential proxy waterfall). |

---

## 2. Ingestion Strategy: Staying Under the Radar

To pull data continuously without suffering IP or account bans, the ingestion architecture adheres to three core tenets:

### A. Request Pacing & Jitter Modeling
Instead of linear polling, requests utilize a **Poisson-like or uniform random jitter distribution**. 
- Minimum delay: `300ms`
- Maximum delay: `900ms`
- Inter-request pause guarantees human-like interaction variance.

### B. Session & User-Agent Rotation
- Rotation of desktop User-Agents (`Chrome 125-127`, `Firefox 128`, `Safari 17.5`) aligned with matching `Sec-Ch-Ua` headers.
- Stateless HTTP client initialization ensures session tracking cookies do not accumulate flags across failed requests.

### C. Proxy Waterfall & Plan B Fallback Cascade
When a primary ingestion source returns `403 Forbidden`, `429 Too Many Requests`, or fails schema validation:
1. **Circuit Breaker Trip**: The service automatically shifts state from `CLOSED` to `OPEN` to prevent burning the client IP further.
2. **Fallback Cascade Execution**:
   - Primary: High-frequency RSS / API.
   - Secondary: Public aggregator feeds.
   - Emergency Plan B: Sandbox mock generator to maintain downstream pipeline availability without zero-data crashes.

---

## 3. Pipeline Resilience & Structural Drift Defense

Target sites frequently alter DOM structures, class names, and API endpoints overnight.

### Architectural Safeguards:
1. **Schema Validation & Structural Isolation**:
   - Incoming payloads pass through a strict parser step (`gofeed` RSS engine / JSON decoder).
   - If parser output contains 0 valid job records, the run is flagged as failed rather than saving corrupted data.
2. **Dead-Letter & Telemetry Metrics**:
   - `/api/metrics` tracks `total_runs`, `successful_runs`, `failed_runs`, `last_error`, and `circuit_breaker_status`.
   - Real-time observability allows rapid detection of markup drift before alerts fire downstream.
3. **Non-Blocking Fault Isolation**:
   - Ingestion failures execute asynchronously or isolated within request contexts without causing Gin API server crashes.

---

## 4. Ethical Boundaries & Terms of Service ("Where to Stop")

Every major job platform specifies prohibitions against automated scraping within its Terms of Service.

### Personal & Technical Boundaries:
1. **Zero CAPTCHA Bypassing**: We do not integrate automated CAPTCHA solvers (2Captcha, Anti-Captcha) to force access past explicit security gates.
2. **No Authenticated Scraping**: We never scrape behind user login walls, store stolen session cookies, or bypass multi-factor authentication.
3. **Robots.txt & Rate Respect**: We enforce rate caps well below human browsing limits and respect explicit disallowed directives.
4. **Transition to Official Feeds**: For commercial production scaling, direct API partnerships (LinkedIn Talent API, Indeed Publisher API) or structured public feeds (RSS/JSON) are the definitive long-term solution.
