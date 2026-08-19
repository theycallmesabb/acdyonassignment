# Architectural & Technical Decisions

## Executive Overview

This document explains the technical decisions behind the **Resilient Job Listing Ingestion Engine**, built with **Golang and Gin**.

The system is designed to ingest job listings from public, low-risk RSS/API sources while handling temporary failures, rate limits, malformed responses, and source changes.

The live demo does not attempt to bypass authentication, CAPTCHA, or anti-bot protections on platforms such as LinkedIn, Indeed, Naukri, or Wellfound.

---

## 1. Detection Surface

Job platforms can identify automated clients using several signals, including:

* Request frequency and repeated request patterns
* Unusual or missing HTTP headers
* User-Agent inconsistencies
* IP reputation and rate limits
* Browser and device fingerprints
* CAPTCHA challenges
* Session and behavioral patterns

The system acknowledges these detection surfaces but does **not** attempt to defeat them.

The live implementation uses direct HTTP requests only against public RSS/API sources and keeps request frequency low.

---

## 2. Ingestion Strategy

The ingestion pipeline uses a simple fallback strategy:

```text
Primary Public Feed
        ↓
   Request fails?
        ↓
Secondary Public Feed
        ↓
   Request fails?
        ↓
   Sandbox Source
        ↓
    Cached Data
```

The HTTP client includes:

* Request timeouts
* Basic request pacing
* Exponential backoff
* Response validation
* Circuit breaker protection

If a source repeatedly returns errors such as `403`, `429`, or `503`, the circuit breaker temporarily stops requests to that source instead of continuously retrying.

This protects both the application and the upstream source.

---

## 3. Resilience

The ingestion layer validates every response before updating the stored job data.

For example, if a feed suddenly changes its format and produces zero valid jobs, the system does not replace the existing data with an empty result.

Instead:

1. The ingestion run is marked as failed.
2. The error is logged.
3. The fallback source is attempted.
4. Previously cached data remains available if all sources fail.

This prevents temporary upstream problems from causing the API to silently return an empty job list.

---

## 4. Ethical Boundaries

The system has clear technical boundaries.

We do **not**:

* Bypass CAPTCHA or access-control systems.
* Scrape authenticated user accounts.
* Circumvent login requirements.
* Use stolen or authenticated sessions.
* Attempt to defeat platform security controls.
* Continuously retry a source after it starts rejecting requests.

The live demo is restricted to public RSS/API feeds and a sandbox source, as required by the assignment's scope guardrail.

If a platform requires authentication, blocks automated access, or does not provide an appropriate public feed, the system stops rather than attempting to circumvent those restrictions.

The architecture is designed so that a blocked or unavailable source can be replaced with an allowed public source without changing the rest of the ingestion pipeline.
