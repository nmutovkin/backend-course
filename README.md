# Backend Engineering: From Sockets to Scale

A comprehensive, language-agnostic backend engineering course that covers every
major topic on the [backend roadmap](https://roadmap.sh/backend). It is built
around **one project that grows through the whole course**:
**Gatepass**, an event-ticketing platform.

> **Concepts first, language second.** Each topic is taught as an idea
> (what problem it solves, how it works underneath, what the trade-offs are),
> and then built for real. We have to write code in *some* language, so the
> reference implementation is written in **Go**. Each module has an
> *Across languages* note that maps the idea to Python, TypeScript/Node,
> Java/Kotlin, C#, and Rust, so you can follow the course in any of them.

---

## Why one project?

Isolated toy examples teach syntax. A system that grows teaches engineering.
Gatepass starts as a raw TCP socket that answers `GET /health`. By the end it
is a containerised, observable, event-driven, horizontally scaled platform
that survives a flash sale of 50,000 concurrent buyers. Every module adds a
**milestone** that forces you to use what you just learned on a real problem:

| Problem in Gatepass                         | What it teaches                                  |
|---------------------------------------------|--------------------------------------------------|
| Two buyers click the same seat at once      | Transactions, isolation levels, locking          |
| A seat is "held" for 10 minutes             | TTLs, background jobs, clocks, distributed state |
| A concert goes on sale at 10:00:00          | Caching, rate limiting, queues, load shedding    |
| Payment provider times out                  | Idempotency, retries, sagas, circuit breakers    |
| "Find jazz events in Berlin this weekend"   | Search engines, geo queries, relevance           |
| Live seat map shows seats disappearing      | WebSockets, SSE, pub/sub fan-out                 |
| Door staff scan tickets with poor signal    | Auth tokens, offline sync, signed payloads       |
| Organisers want sales dashboards            | OLAP, time-series data, GraphQL                  |

See [`PROJECT.md`](./PROJECT.md) for the full specification and per-module milestones.

---

## Course map

The course is split into **6 parts** and **33 modules** (00–32).
Full breakdown with every sub-topic: [`SYLLABUS.md`](./SYLLABUS.md).

### Part I: Foundations
| #  | Module                                  | Gatepass milestone                                  |
|----|-----------------------------------------|-----------------------------------------------------|
| 00 | Setup & Go Primer                       | Environment ready; Gatepass tests pass              |
| 01 | How the Internet Works                  | Hand-craft HTTP requests; trace DNS → TCP → TLS     |
| 02 | Operating Systems & Linux               | `gatepass` CLI, signals, graceful shutdown           |
| 03 | The Programming Language                | In-memory domain model: events, venues, seats       |
| 04 | Git & Collaboration                     | Repo, branching model, PR template, conventions     |

### Part II: Building the Core Service
| #  | Module                                  | Gatepass milestone                                  |
|----|-----------------------------------------|-----------------------------------------------------|
| 05 | HTTP & Web Servers from Scratch         | HTTP/1.1 server on raw sockets, then a real router  |
| 06 | Data Formats, Encoding & the Tricky Basics | Money, time zones, IDs, Unicode done right       |
| 07 | Relational Databases & SQL              | PostgreSQL schema; double-booking-proof reservations|
| 08 | API Design (REST)                       | OpenAPI-first public API, pagination, idempotency   |
| 09 | Application Architecture & Design       | Refactor to a hexagonal, modular monolith           |
| 10 | Testing                                 | Unit → integration → contract → concurrency tests   |
| 11 | Authentication & Authorization          | Accounts, sessions, JWT, OAuth/OIDC, RBAC           |
| 12 | Web Security                            | Threat model, OWASP hardening, secrets, TLS         |

### Part III: Data at Scale
| #  | Module                                  | Gatepass milestone                                  |
|----|-----------------------------------------|-----------------------------------------------------|
| 13 | Caching                                 | HTTP caching, Redis cache-aside, stampede protection|
| 14 | NoSQL & Beyond Relational               | Redis holds, document store, analytics in ClickHouse|
| 15 | Search Engines                          | Full-text, faceted, geo and autocomplete search     |
| 16 | File Storage & Media                    | Poster uploads, thumbnails, PDF/QR tickets          |
| 17 | Scaling Databases                       | Read replicas, partitioning, zero-downtime migrations|

### Part IV: Asynchronous & Distributed Systems
| #  | Module                                  | Gatepass milestone                                  |
|----|-----------------------------------------|-----------------------------------------------------|
| 18 | Concurrency, Background Jobs & Reliability | Hold expiry worker, outbox, retries, idempotency |
| 19 | Message Brokers & Event-Driven Architecture | Order saga over Kafka/RabbitMQ                  |
| 20 | Alternative API Styles                  | gRPC payments, GraphQL dashboard, webhooks          |
| 21 | Real-Time Communication                 | Live seat map (WebSockets), queue position (SSE)    |
| 22 | Distributed Systems Fundamentals        | Clocks, consensus, leader election for schedulers   |

### Part V: Production Engineering
| #  | Module                                  | Gatepass milestone                                  |
|----|-----------------------------------------|-----------------------------------------------------|
| 23 | Containers, Virtualization & Orchestration | Docker images, Compose, Kubernetes manifests     |
| 24 | Web Servers, Proxies & Load Balancing   | Nginx/Caddy reverse proxy, TLS, L7 load balancing   |
| 25 | CI/CD & Infrastructure as Code          | Pipeline, Terraform, canary deploys, feature flags  |
| 26 | Cloud Fundamentals                      | Deploy to a cloud with managed services on a budget |
| 27 | Observability                           | Logs, metrics, traces, SLOs and alerting            |
| 28 | Architectural Patterns                  | Strangler-fig extraction of services, ADRs          |
| 29 | Building for Scale & Resilience         | Survive a 50k-user flash sale                       |

### Part VI: Synthesis
| #  | Module                                  | Gatepass milestone                                  |
|----|-----------------------------------------|-----------------------------------------------------|
| 30 | System Design                           | Design doc: Gatepass goes global                    |
| 31 | AI-Powered Backends *(elective)*        | Semantic search & RAG support assistant             |
| 32 | Capstone: Launch Day                    | Game day, runbooks, final review & demo             |

Coverage check against roadmap.sh: [`ROADMAP_COVERAGE.md`](./ROADMAP_COVERAGE.md).

---

## How deep we go

About 450 hours. Each sub-topic in the syllabus has a depth tag:

- **`L3` Internals:** built from scratch first, then the real tool. Reserved for what backend engineers get paged for: TCP/HTTP, SQL & transactions, concurrency & idempotency, caching, auth, API design, observability, resilience.
- **`L2` Production:** used for real in Gatepass, with failure labs. Most tools: Redis, Kafka, gRPC, GraphQL, Docker, Kubernetes (developer side), Terraform, cloud.
- **`L1` Awareness:** what it is, when to choose it, trade-offs. Examples: SOAP, SAML, Cassandra, Paxos, service mesh.

Target outcome: a strong mid-level engineer heading toward senior. Details and the list of from-scratch builds are in [`SYLLABUS.md`](./SYLLABUS.md#depth-levels).

## Repository layout

```
modules/NN-slug/README.md   the lessons (written in order; see docs/LEDGER.md for status)
gatepass/                   the project's reference implementation (one Go module)
docs/AUTHORING.md           how modules are written and kept consistent
docs/LEDGER.md              what has been built, defined, decided and promised so far
```

## How every module is structured

Each module under `modules/NN-slug/` follows the same shape:

1. **Why this exists**: the problem, told through Gatepass.
2. **Concepts**: language-agnostic explanation, diagrams, trade-offs.
3. **Under the hood**: how it actually works (protocol bytes, data structures, algorithms).
4. **Build it**: the reference implementation in Go, step by step.
5. **Across languages**: the same idea in Python, TypeScript, Java/Kotlin, C#, Rust.
6. **Labs**: small, focused exercises.
7. **Project milestone**: the Gatepass increment, with acceptance criteria.
8. **Checkpoint**: questions you should be able to answer before moving on.
9. **Further reading**: a short, curated list.

## Prerequisites

- Comfortable writing basic programs in at least one language. Go is taught
  in the course (Module 00 primer, Module 03 in depth).
- A machine (or VM) running Linux or macOS, or Windows with WSL2.
- Docker from Module 07 on. Module 00 walks through installing everything.

## Self-contained by design

Everything you need to finish every module (concepts, protocol details, tool
installation, lab steps, expected output, and answers to every question) is
in the course itself. Books, RFCs and articles are listed only as optional
further reading.

## Why Go as the reference language?

Go is small enough to learn alongside the course, close enough to the metal
that sockets, threads and memory stay visible, and widely used for backend
infrastructure. Its standard library has enough in it (HTTP, TLS, SQL, JSON)
that we can build things *before* reaching for frameworks. That matters
because the course builds things from first principles before using the
industrial version. **None of the concepts depend on Go.** Swap it for any
language listed above.

## Core references used throughout

- *Designing Data-Intensive Applications*, Martin Kleppmann
- *Computer Networking: A Top-Down Approach*, Kurose & Ross
- *Database Internals*, Alex Petrov
- *Release It!*, Michael Nygard
- *Site Reliability Engineering*, Google
- *The Twelve-Factor App*, 12factor.net
- OWASP Top 10 & Cheat Sheet Series
