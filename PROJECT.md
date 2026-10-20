# Gatepass: The Course Project

Gatepass is an **event-ticketing platform**. Organizers publish events at
venues with seat maps. Fans search, reserve seats, pay, and receive tickets.
Door staff scan tickets at the entrance.

Ticketing is a good course project because it is hard in all the ways
backends are hard:

- **Correctness under contention.** Thousands of people want the same seat at the same second.
- **Time.** Holds expire, events have time zones, on-sales start at an exact instant.
- **Money.** Payments fail, time out, and get retried. You must never charge twice or sell a seat twice.
- **Spiky load.** Traffic is ~0 most of the time and 100× normal at 10:00:00 on sale day.
- **Many read patterns.** Browsing, search, live seat maps, analytics, check-in.
- **Many actors.** Fans, organizers, staff, admins, and partner integrations.

---

## 1. Product overview

### Actors
| Actor        | Description                                                         |
|--------------|---------------------------------------------------------------------|
| Fan          | Browses, searches, buys tickets, manages orders                     |
| Organizer    | Creates events, configures pricing, views sales, issues refunds     |
| Staff        | Scans tickets at the door (mobile app, unreliable network)          |
| Admin        | Platform operator: moderation, support, configuration               |
| Integrator   | Organizer's own systems via API keys and webhooks                   |

### Core flows
1. **Publish:** organizer creates a venue (sections, rows, seats), an event, price tiers, and an on-sale time.
2. **Discover:** fan browses and searches events by text, city, date, genre, price, and distance.
3. **Reserve:** fan picks seats (or general admission quantity). Seats are **held for 10 minutes**.
4. **Checkout:** fan pays; on success, order is confirmed and tickets are issued; on failure or timeout, holds are released.
5. **Deliver:** fan receives email confirmation and PDF tickets with signed QR codes.
6. **Check-in:** staff scan QR codes; a ticket can be admitted exactly once, even when scanners are offline for part of the night.
7. **Refund:** organizer or admin refunds an order; seats return to inventory.
8. **Analyze:** organizer sees live sales dashboards.

### Non-functional targets (final state)
| Property          | Target                                                                 |
|-------------------|------------------------------------------------------------------------|
| Correctness       | Zero double-sold seats, zero double charges (proved by tests)          |
| Checkout SLO      | 99.9% of checkouts succeed within 2 s (p99) during normal operation    |
| Flash sale        | 50,000 concurrent users at on-sale without breaching the checkout SLO  |
| Browse latency    | p99 < 200 ms for event pages (cached)                                  |
| Availability      | 99.9% monthly for public API                                           |
| Durability        | RPO ≤ 1 min, RTO ≤ 30 min                                              |
| Security          | OWASP Top 10 addressed; no card data stored (payment provider tokens)  |

---

## 2. Bounded contexts

```
┌───────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐
│ Identity  │  │  Catalog   │  │ Inventory  │  │  Ordering  │  │ Ticketing  │
│ users,    │  │ venues,    │  │ seats,     │  │ carts,     │  │ tickets,   │
│ roles,    │  │ events,    │  │ holds,     │  │ orders,    │  │ QR codes,  │
│ sessions  │  │ pricing    │  │ availability│ │ saga       │  │ check-in   │
└───────────┘  └────────────┘  └────────────┘  └────────────┘  └────────────┘
┌───────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐
│ Payments  │  │  Search    │  │Notifications│ │ Analytics  │
│ provider  │  │ index,     │  │ email, SMS, │ │ sales,     │
│ adapter   │  │ queries    │  │ webhooks    │ │ dashboards │
└───────────┘  └────────────┘  └────────────┘  └────────────┘
```

## 3. Architecture evolution

Gatepass deliberately starts simple and gets more complex only when a module
gives a reason to.

| Stage | Modules | Shape                                                                  |
|-------|---------|------------------------------------------------------------------------|
| A     | 01–06   | Single binary, in-memory data, hand-rolled HTTP → stdlib server        |
| B     | 07–12   | Modular monolith + PostgreSQL; authenticated, tested, secured          |
| C     | 13–17   | + Redis, search engine, object storage, ClickHouse, read replicas      |
| D     | 18–22   | + workers, message broker, Payments service (gRPC), real-time gateway  |
| E     | 23–29   | Containerised, behind a proxy, CI/CD, cloud-deployed, observable, scaled; Notifications & Search Indexer extracted |
| F     | 30–32   | Designed for global scale; AI features; production-launched            |

### Final architecture (Stage E/F)

```
                     ┌──────────┐
   Fans / Staff ───▶ │   CDN    │──▶ static assets, cached event pages
                     └────┬─────┘
                          ▼
                 ┌─────────────────┐
                 │ Reverse proxy / │  TLS, rate limiting, waiting room
                 │  load balancer  │
                 └──┬───────────┬──┘
                    ▼           ▼
         ┌────────────────┐  ┌───────────────┐
         │ Gatepass API   │  │ Realtime GW   │ WebSockets / SSE
         │ (modular       │  └──────┬────────┘
         │  monolith)     │         │ pub/sub
         └─┬──┬───┬───┬───┘         │
   gRPC ◀──┘  │   │   └────────┐    │
 ┌────────┐   │   │            ▼    ▼
 │Payments│   │   │        ┌──────────┐
 │service │   │   │        │  Redis   │ cache, holds, rate limits, pub/sub
 └───┬────┘   │   │        └──────────┘
     ▼        ▼   ▼
 fake PSP ┌────────┐ ┌───────────┐       ┌──────────────┐
          │Postgres│ │ Object    │       │ Kafka /      │
          │+replica│ │ storage   │       │ RabbitMQ     │
          └───┬────┘ └───────────┘       └──┬───┬───┬───┘
              │ outbox ─────────────────────┘   │   │
                         ▼                      ▼   ▼
              ┌──────────────┐  ┌────────────┐ ┌──────────┐
              │ Notifications│  │  Search    │ │Analytics │
              │  service     │  │  indexer → │ │→ClickHouse│
              └──────────────┘  │ OpenSearch │ └──────────┘
                                └────────────┘
    Telemetry: OpenTelemetry → Collector → Prometheus / Loki / Tempo → Grafana
```

## 4. Repository layout (target)

```
gatepass/
├── cmd/                 # entry points: api, worker, payments, realtime, cli
├── internal/
│   ├── identity/
│   ├── catalog/
│   ├── inventory/
│   ├── ordering/
│   ├── ticketing/
│   ├── payments/
│   ├── search/
│   ├── notifications/
│   ├── analytics/
│   └── platform/        # shared infra: http, db, messaging, telemetry, config
├── api/                 # openapi.yaml, .proto files, graphql schema, event schemas
├── migrations/
├── deploy/              # Dockerfiles, compose, k8s, terraform
├── loadtests/           # k6 scripts
├── docs/                # ADRs, threat model, runbooks, design docs
└── tests/               # e2e and contract tests
```

---

## 5. Milestones

Each milestone builds on the previous one. Acceptance criteria are what a
reviewer checks before the milestone counts as done.

### Part I: Foundations

**M01: `gpcurl`** *(Module 01)*
- Raw TCP client that sends a hand-written HTTP/1.1 request and parses the response.
- Prints DNS, connect, TLS handshake and TTFB timings.
- ✅ Works against `http://` and `https://` hosts; handles chunked responses.

**M02: `gatepass` CLI skeleton** *(Module 02)*
- Subcommands (`serve`, `migrate`, `seed`), config from flags and env vars.
- Graceful shutdown on `SIGTERM`/`SIGINT` with a drain timeout.
- ✅ Meaningful exit codes; `kill -TERM` completes in-flight work.

**M03: Domain model** *(Module 03)*
- Types: `Venue`, `Section`, `Seat`, `Event`, `PriceTier`, `Hold`, `Order`, `Ticket`.
- Rules: a seat has at most one active hold or sale; holds expire; an order total equals the sum of its line items.
- ✅ Unit tests for every rule; concurrent access is safe (race detector clean).

**M04: Repository & workflow** *(Module 04)*
- Monorepo layout, PR template, CODEOWNERS, conventional commits, pre-commit hooks.
- ✅ First ADR: "Modular monolith first".

### Part II: Core Service

**M05: HTTP server** *(Module 05)*
- v0: HTTP/1.1 server on raw sockets serving `/health` and `/events`.
- v1: stdlib server + router; middleware for request IDs, logging, panic recovery, timeouts.
- ✅ Survives slowloris; body size limits enforced.

**M06: Getting the basics right** *(Module 06)*
- `Money` (minor units + currency), event times in UTC + venue IANA zone, UUIDv7 IDs.
- ✅ Tests for DST transitions, currency rounding, Unicode event names.

**M07: PostgreSQL** *(Module 07)*
- Schema and migrations for all core tables, with constraints that enforce invariants.
- Seat reservation in a transaction (`SELECT … FOR UPDATE` or a unique constraint; trade-offs documented).
- ✅ Keyset pagination; `EXPLAIN ANALYZE` evidence for critical queries.

**M08: Public REST API** *(Module 08)*
- OpenAPI 3.1 spec written first; server implements it.
- Problem Details errors, cursor pagination, filtering, `Idempotency-Key` on `POST /orders`.
- ✅ Spec-compliance test passes; retried order creation never creates duplicates.

**M09: Modular monolith** *(Module 09)*
- Code split into bounded contexts with ports & adapters.
- ✅ Import rules enforced by a linter; no context touches another's tables.

**M10: Test suite** *(Module 10)*
- Unit, integration (Testcontainers), API e2e, and contract tests.
- ✅ 1,000 parallel reservations for one seat → exactly one succeeds; fuzz tests on input parsing.

**M11: Identity** *(Module 11)*
- Fan accounts with Argon2id + session cookies; OIDC social login.
- Organizer/staff/admin roles (RBAC) with per-organizer tenancy.
- Short-lived JWTs for the scanner app; API keys for integrators.
- ✅ Authorization tests for every endpoint; IDOR tests pass.

**M12: Security hardening** *(Module 12)*
- Threat model (STRIDE) in `docs/`; security headers, CORS, CSRF.
- Login rate limits; secrets in a secret manager; dependency scanning.
- ✅ The deliberately vulnerable branch's exploits all fail on main.

### Part III: Data at Scale

**M13: Caching** *(Module 13)*
- `Cache-Control`/`ETag` on public pages; Redis cache-aside for event details.
- Request coalescing for hot keys; invalidation on event update.
- ✅ Load test shows hit ratio > 95% and no stampede on cache expiry.

**M14: Polyglot persistence** *(Module 14)*
- Seat holds in Redis with TTL (Postgres remains source of truth for sales).
- Sales events streamed to ClickHouse for analytics.
- ✅ Written ADR on each store choice.

**M15: Search** *(Module 15)*
- OpenSearch index: full text, facets, geo radius, autocomplete, synonyms.
- ✅ Relevance test set passes; zero-downtime reindex via aliases.

**M16: Files & tickets** *(Module 16)*
- Poster uploads via presigned URLs to MinIO/S3; async thumbnails.
- PDF tickets with signed (Ed25519) QR payloads; signed download links.
- ✅ 1 GB upload never exceeds 50 MB server memory; a tampered QR is rejected.

**M17: Database scaling** *(Module 17)*
- Read replicas for browse; read-your-writes for a fan's own orders.
- Time-partitioned `orders`; a zero-downtime expand/contract migration; restore drill.
- ✅ Documented RPO/RTO measured from the drill.

### Part IV: Async & Distributed

**M18: Background work** *(Module 18)*
- Workers: hold expiry, emails, ticket generation; transactional outbox.
- ✅ Kill workers mid-job: no lost or duplicated emails/tickets.

**M19: Event-driven order saga** *(Module 19)*
- Order saga: reserve → pay → issue, with compensations (release, refund).
- Domain events on Kafka/RabbitMQ with versioned schemas; DLQ handling.
- ✅ Chaos test: kill any participant at any step; saga always ends consistent.

**M20: gRPC, GraphQL, webhooks** *(Module 20)*
- Payments service over gRPC with deadlines; fake PSP with random failures/latency.
- GraphQL organizer dashboard API with DataLoader and complexity limits.
- Signed outgoing webhooks with retries; verified incoming PSP webhooks.
- ✅ Replayed webhook is rejected; deep GraphQL query is rejected.

**M21: Real-time** *(Module 21)*
- Live seat map over WebSockets, fan-out via Redis pub/sub across nodes.
- SSE stream of waiting-room position.
- ✅ 10,000 concurrent sockets on a single node; reconnect resumes without missed updates.

**M22: Distributed coordination** *(Module 22)*
- Leader election for schedulers; fencing tokens on inventory writes.
- ✅ Network-partition test: no double processing; consistency guarantees documented.

### Part V: Production

**M23: Containers** *(Module 23)*
- Multi-stage, distroless images; `docker compose up` brings up the full stack.
- Kubernetes manifests (probes, limits, HPA) running on kind/k3d.
- ✅ API image < 30 MB; rolling update with zero failed requests.

**M24: Edge & load balancing** *(Module 24)*
- Caddy/Nginx: TLS, HTTP/2, compression, routing API vs WebSocket traffic, health checks.
- ✅ Killing one of three API instances under load causes no user-visible errors.

**M25: CI/CD & IaC** *(Module 25)*
- Pipeline: lint → test → build → scan → push → deploy; Terraform for infrastructure.
- Canary deployment with automated rollback; feature flags.
- ✅ A deliberately broken release is rolled back automatically.

**M26: Cloud deployment** *(Module 26)*
- Managed Postgres, Redis, object storage, broker; private networking; least-privilege IAM.
- ✅ Monthly cost estimate under budget; budget alerts configured.

**M27: Observability** *(Module 27)*
- OpenTelemetry traces across HTTP, gRPC and messages; structured logs with trace IDs.
- RED dashboards; checkout SLO with burn-rate alerts; runbooks.
- ✅ Given an injected fault, the on-call can find the root cause from telemetry alone.

**M28: Service extraction** *(Module 28)*
- Notifications and Search Indexer extracted via strangler fig.
- ✅ ADRs for extracted vs retained contexts; no shared database between services.

**M29: Flash sale** *(Module 29)*
- k6 scenario: 50,000 VUs at on-sale time.
- Virtual waiting room, token-bucket rate limiting, circuit breaker on PSP, load shedding, graceful search degradation.
- ✅ Checkout SLO holds; zero oversells; report with bottleneck analysis.

### Part VI: Synthesis

**M30: Gatepass Global** *(Module 30)*: design doc for multi-region, data residency, global on-sales.

**M31: AI features** *(Module 31, elective)*: semantic event search, RAG support assistant with injection defenses and evals.

**M32: Launch day** *(Module 32)*: production readiness review, game day with scripted failures, postmortem, final demo.
