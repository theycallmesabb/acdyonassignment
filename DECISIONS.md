# Architectural & Technical Decisions (`DECISIONS.md`)

## Executive Overview
This document summarizes the core technical decisions for the **Resilient Job Listing Ingestion Engine** built in **Golang (Gin)**, addressing anti-bot detection mitigation, ingestion strategy, system resilience, and compliance boundaries.

---

## 1. Detection Surface & Countermeasures
High-volume platforms (LinkedIn, Indeed, Naukri, Wellfound) deploy multi-layer bot detection. Our design accounts for these specific vectors:

* **HTTP Header Alignment**: Standard automated HTTP clients lack browser-native headers. We emulate full Chrome/macOS header ordering (`Sec-Ch-Ua`, `Sec-Fetch-Dest`, `Sec-Fetch-Mode`, `Sec-Fetch-Site`, `Accept-Language`).
* **User-Agent Pool Rotation**: Randomly rotates valid desktop browser User-Agents (`Chrome 125-127`, `Firefox 128`, `Safari 17.5`) per request batch.
* **Temporal Jittering**: Replaces metronomic sub-second polling with uniform random delays (`300ms–900ms`) to break statistical timing signatures.
* **Headless Browser Avoidance**: Uses direct HTTP payload ingestion to avoid WebGL/Canvas fingerprinting and Chrome DevTools Protocol (`navigator.webdriver`) exposure.

---

## 2. Ingestion Strategy & Plan B Cascade
* **Pacing & Identity**: Stateless request handling prevents accumulating tracking flags across retries while maintaining low-frequency request distribution.
* **Circuit Breaker Pattern**: Tracks block status codes (`403`, `429`, `503`). Tripping thresholds triggers a shift to `OPEN` state for 30s to protect IP reputation.
* **Plan B Fallback Cascade**: If the primary source fails or blocks, traffic seamlessly shifts to:
  1. Primary Feed (High-frequency RSS/Public Feed)
  2. Secondary Feed (Backup Aggregator RSS)
  3. Sandbox Mock Engine (Zero-downtime mock fallback to ensure downstream pipelines never crash).

---

## 3. Resilience & Markup Drift Defense
* **Non-Zero Record Validation**: Ingested payloads pass strict schema parsing (`gofeed` XML parser). If markup changes lead to 0 parsed records, the run is logged as failed rather than saving blank data.
* **Observability Telemetry**: `/api/metrics` tracks total runs, success/failure counts, active source, and circuit breaker status for immediate drift visibility.

---

## 4. Ethical Boundaries ("Where We Stop")
* **No CAPTCHA Bypassing**: We do not use automated CAPTCHA solving services.
* **No Authenticated Scraping**: We do not scrape behind user logins or harvest gated data.
* **Public & Allowed Feeds Only**: Ingestion is restricted to low-risk public RSS/API endpoints and sandboxes in full compliance with scope guardrails.
