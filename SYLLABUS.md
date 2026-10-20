# Syllabus

Every module lists its **topics** (grouped into numbered sub-sections),
its **labs**, an **Across languages** note, and its **Gatepass milestone**.
Full milestone specs with acceptance criteria are in [`PROJECT.md`](./PROJECT.md).

Estimated effort: ~400–500 hours total. Each module takes 8–20 hours.

---

# Part I: Foundations

## 01. How the Internet Works

**Goal:** know what actually happens between "user presses Enter" and
"server receives bytes", well enough to debug it.

1. **Networking models**
   - OSI vs TCP/IP models; what each layer is responsible for
   - Encapsulation: frames, packets, segments, messages
2. **Addressing & routing**
   - IPv4, IPv6, subnets, CIDR notation, private ranges, NAT
   - Routing, gateways, BGP at a glance, traceroute
   - MAC addresses, ARP
3. **Transport**
   - TCP: three-way handshake, sequence numbers, flow control, congestion control, retransmission, `TIME_WAIT`, Nagle
   - UDP: when unreliable is better (DNS, video, QUIC)
   - Ports, sockets, ephemeral ports, the 5-tuple
4. **DNS**
   - Hierarchy: root, TLD, authoritative, recursive resolvers
   - Record types: A, AAAA, CNAME, MX, TXT, NS, SRV, CAA
   - TTLs, caching, propagation myths, DNS-based load balancing
   - Tools: `dig`, `nslookup`, `host`
5. **HTTP**
   - HTTP/1.0 → 1.1 (keep-alive, chunked) → HTTP/2 (multiplexing, HPACK) → HTTP/3 (QUIC)
   - Anatomy of a request/response: method, path, headers, body
   - Methods, status code classes, common headers
6. **TLS & HTTPS (intro)**
   - Why encryption + authentication; handshake overview; certificates (deep dive in Module 12)
7. **Browsers & the web**
   - What a browser does: DNS, connection, rendering, same-origin policy
   - Domain names, registrars, hosting types (shared, VPS, cloud, serverless)
   - CDN overview

**Labs:** capture a full page load in Wireshark; write an HTTP request by hand with `nc`/`telnet`; resolve a domain recursively with `dig +trace`; measure TTFB with `curl -w`.

**Across languages:** every language wraps the same BSD socket API: `net` (Go), `socket` (Python), `net` (Node), `java.net`/NIO, `System.Net.Sockets`, `std::net` (Rust).

**Gatepass M01:** a tiny `gpcurl` tool that opens a TCP socket, speaks raw HTTP/1.1, and prints timing per phase (DNS, connect, TLS, TTFB).

---

## 02. Operating Systems & Linux

**Goal:** understand the machine your server runs on.

1. **The terminal**
   - Shell basics, pipes, redirection, exit codes, job control
   - Essential tools: `grep`, `sed`, `awk`, `find`, `xargs`, `jq`, `curl`, `ssh`, `tmux`
   - Writing robust shell scripts (`set -euo pipefail`)
2. **Processes & threads**
   - Process lifecycle, `fork`/`exec`, PIDs, parent/child, zombies, orphans
   - Threads vs processes; context switching; scheduling basics
   - Signals: `SIGTERM`, `SIGINT`, `SIGKILL`, `SIGHUP`; graceful shutdown
3. **Memory**
   - Virtual memory, pages, stack vs heap, memory-mapped files
   - OOM killer, swap, RSS vs VSZ
4. **Files & I/O**
   - Everything is a file; file descriptors; limits (`ulimit`)
   - Blocking vs non-blocking I/O; `select`/`poll`/`epoll`/`kqueue`/`io_uring`
   - Buffering, `fsync`, durability
5. **Filesystem & permissions**
   - FHS layout, users/groups, permissions, ownership, `sudo`
   - Inodes, hard/soft links
6. **Networking on the host**
   - `ss`/`netstat`, `ip`, `lsof`, firewall basics (`iptables`/`nftables`/`ufw`)
7. **Services & system management**
   - `systemd` units, journald, cron
   - Package managers, environment variables
   - POSIX basics and why portability matters

**Labs:** find which process holds a port; watch syscalls with `strace`; trigger and observe the OOM killer in a container; write a systemd unit.

**Across languages:** signal handling (`os/signal`, `signal` module, `process.on('SIGTERM')`, `Runtime.addShutdownHook`, `IHostApplicationLifetime`, `tokio::signal`).

**Gatepass M02:** `gatepass` CLI skeleton with subcommands, config via env vars and flags, structured exit codes, and graceful shutdown on `SIGTERM`.

---

## 03. The Programming Language

**Goal:** be productive in your backend language and understand what its runtime does for you.

1. **Language fundamentals**
   - Types, structs/classes, interfaces/traits, generics
   - Error handling models: exceptions vs error values vs `Result`
   - Modules, packages, visibility
2. **Tooling**
   - Dependency management & lockfiles; reproducible builds
   - Formatters, linters, static analysis
   - Build, run, cross-compile, debug
3. **Concurrency models**
   - OS threads, green threads/goroutines, async/await + event loop, actors
   - CPU-bound vs I/O-bound work; the GIL and its equivalents
   - Channels, futures/promises, cancellation and context propagation
4. **Memory management**
   - Manual, reference counting, tracing GC, ownership/borrowing
   - Allocation costs, escape analysis, GC pauses
5. **The standard library for backends**
   - HTTP, JSON, time, crypto, I/O, logging
6. **Profiling & performance basics**
   - CPU and memory profiling, benchmarks, flame graphs
7. **Idioms & code quality**
   - Naming, small interfaces, composition over inheritance, documentation comments

**Labs:** implement the same concurrent URL fetcher using threads, async, and goroutines/channels; profile a deliberately slow function.

**Across languages:** comparison table of Go, Python, TypeScript/Node, Java/Kotlin, C#, Rust (runtime, concurrency model, typing, ecosystem, deploy artifact).

**Gatepass M03:** in-memory domain model (`Venue`, `Seat`, `Event`, `Order`, `Ticket`) with business rules (no overselling, hold expiry) and unit tests.

---

## 04. Git & Collaboration

**Goal:** use version control like a professional team member.

1. **Git internals**
   - Objects (blob, tree, commit, tag), refs, HEAD, the index
2. **Everyday Git**
   - Branching, merging, rebasing, cherry-picking, stashing, bisect, reflog
   - Resolving conflicts; rewriting history safely
3. **Workflows**
   - Trunk-based, GitHub Flow, GitFlow; trade-offs
   - Monorepo vs polyrepo
4. **Repository hosting**
   - GitHub, GitLab, Bitbucket; pull/merge requests; code owners
   - Code review: what to look for, how to give and receive feedback
5. **Conventions**
   - Conventional commits, semantic versioning, changelogs
   - `.gitignore`, `.gitattributes`, pre-commit hooks
   - Documentation: README, ADRs, CONTRIBUTING

**Labs:** recover a "lost" commit via reflog; find a bug with `git bisect`; resolve a three-way conflict.

**Gatepass M04:** the Gatepass monorepo: layout, branch protection, PR template, commit conventions, pre-commit hooks, first ADR.

---

# Part II: Building the Core Service

## 05. HTTP & Web Servers from Scratch

**Goal:** know exactly what a web framework does by building one.

1. **A server on raw sockets**
   - Accept loop, connection handling, thread-per-connection vs event loop
   - Parsing HTTP/1.1 requests; writing responses; keep-alive; chunked encoding
2. **Request lifecycle**
   - Routing (trie/radix routers), path params, query strings
   - Middleware/interceptor chains: logging, recovery, timing
   - Request context, deadlines, cancellation
3. **HTTP semantics in depth**
   - Methods & safety/idempotency, status codes in practice
   - Headers: content negotiation, `Content-Type`, `Accept`, compression
   - Cookies: attributes (`HttpOnly`, `Secure`, `SameSite`, `Domain`, `Path`)
   - Redirects, range requests, `Expect: 100-continue`
4. **Serving the frontend**
   - HTML, CSS, JavaScript basics: enough to understand what you serve
   - Static files, SSR vs CSR vs SSG, templates
   - Forms, `multipart/form-data`, `application/x-www-form-urlencoded`
5. **Frameworks**
   - Anatomy of a framework: router, binding, validation, DI
   - Micro vs batteries-included frameworks
6. **Server hardening basics**
   - Timeouts (read, write, idle, header), body size limits, slowloris

**Labs:** build a static file server; implement a radix-tree router; break your own server with slowloris.

**Across languages:** `net/http` + chi/echo (Go), FastAPI/Django/Flask, Express/Fastify/NestJS, Spring Boot/Ktor, ASP.NET Core, Axum/Actix.

**Gatepass M05:** Gatepass API v0 on raw sockets (`/health`, `/events`), then migrated to the stdlib HTTP server and a router with middleware.

---

## 06. Data Formats, Encoding & the Tricky Basics

**Goal:** stop shipping the bugs that hit every backend sooner or later.

1. **Serialization formats**
   - JSON (and its edge cases: big numbers, `null` vs missing), XML, YAML, TOML, CSV
   - Binary formats: Protocol Buffers, Avro, MessagePack, CBOR
   - Schema vs schemaless; forward/backward compatibility
2. **Text & encoding**
   - ASCII, Unicode, code points, UTF-8/16; normalization; grapheme clusters
   - Base64, URL encoding, percent-encoding
3. **Time**
   - UTC vs local time, time zones, the IANA tz database, DST traps
   - ISO 8601 / RFC 3339; Unix timestamps; monotonic vs wall clocks
4. **Money & numbers**
   - Why floats are wrong for money; minor units and decimals; currency; rounding
5. **Identifiers**
   - Auto-increment vs UUID v4 vs UUID v7 vs ULID vs Snowflake; enumeration attacks
6. **Validation**
   - Parse, don't validate; boundary validation; schemas (JSON Schema)

**Labs:** reproduce a DST bug; round-trip a Protobuf message across schema versions; find a float rounding bug in a price calculation.

**Gatepass M06:** `Money` and `Instant` value types, event times stored in UTC with venue time zones, UUIDv7 IDs, JSON contracts with validation.

---

## 07. Relational Databases & SQL

**Goal:** model data correctly and make the database do the hard work safely.

1. **The relational model**
   - Tables, rows, keys (primary, foreign, natural, surrogate), constraints
   - Normalization (1NF–BCNF) and deliberate denormalization
   - ER diagrams and data modeling
2. **SQL**
   - DDL, DML, `JOIN`s, aggregates, `GROUP BY`/`HAVING`, subqueries
   - CTEs, recursive queries, window functions, `UPSERT`
   - Views, materialized views, stored procedures, triggers (and when not to use them)
3. **Transactions**
   - ACID properties
   - Isolation levels: read uncommitted → serializable; anomalies (dirty read, non-repeatable read, phantom, write skew, lost update)
   - MVCC; locks (row, table, advisory); `SELECT … FOR UPDATE`, `SKIP LOCKED`
   - Optimistic vs pessimistic concurrency control; deadlocks
4. **Indexes & performance**
   - B-tree, hash, GIN/GiST, BRIN; composite, partial, covering indexes
   - `EXPLAIN ANALYZE`, query planner, statistics
   - N+1 queries, pagination performance (offset vs keyset)
5. **Working from code**
   - Drivers, connection pooling (and PgBouncer), prepared statements
   - Raw SQL vs query builders vs ORMs: trade-offs
   - Schema migrations: tooling, versioning, rollbacks
6. **Database internals (intro)**
   - Pages, WAL, checkpoints, vacuum
7. **Other relational engines**
   - PostgreSQL (reference), MySQL/MariaDB, SQLite, SQL Server, Oracle: differences that matter

**Labs:** reproduce each isolation anomaly in two `psql` sessions; speed up a slow query 100× with an index; write a recursive CTE for a venue section tree.

**Across languages:** `database/sql` + pgx/sqlc (Go), SQLAlchemy/psycopg, Prisma/Drizzle/Knex, JDBC/JPA/jOOQ, EF Core/Dapper, SQLx/Diesel.

**Gatepass M07:** PostgreSQL schema for venues, seat maps, events, price tiers, orders, tickets; migrations; seat reservation that is provably free of double-booking.

---

## 08. API Design (REST)

**Goal:** design APIs that clients love and that survive change.

1. **REST principles**
   - Resources, representations, uniform interface, statelessness
   - Richardson Maturity Model; HATEOAS
2. **Practical design**
   - URL design, nested vs flat resources, naming
   - Correct status codes; error format (RFC 9457 Problem Details)
   - Pagination (offset, cursor/keyset), filtering, sorting, sparse fieldsets
   - Bulk operations, long-running operations (`202 Accepted` + status resource)
3. **Idempotency & concurrency**
   - Idempotent methods; `Idempotency-Key` header
   - Optimistic concurrency with `ETag`/`If-Match`
4. **Versioning & evolution**
   - URL vs header versioning, additive change, deprecation & sunset headers
5. **Specifications & tooling**
   - OpenAPI, API-first design, code generation, mock servers
   - JSON:API and other conventions
   - Documentation, SDKs, developer experience
6. **API styles overview**
   - REST vs RPC vs GraphQL vs gRPC vs SOAP (deep dive in Module 20)

**Labs:** review and redesign a badly designed API; generate a client from OpenAPI; implement cursor pagination.

**Gatepass M08:** OpenAPI 3.1 spec for the public API, implemented server-side; cursor pagination, filtering, Problem Details errors, `Idempotency-Key` on order creation.

---

## 09. Application Architecture & Design Principles

**Goal:** write code that stays maintainable as Gatepass grows.

1. **Principles**
   - SOLID, DRY, KISS, YAGNI, separation of concerns, law of Demeter
   - Coupling & cohesion, composition over inheritance
2. **Code organization**
   - Layered architecture, hexagonal (ports & adapters), clean/onion architecture
   - Package by feature vs by layer; modular monolith boundaries
   - Dependency injection (manual and with containers)
3. **Design patterns for backends**
   - GoF essentials: strategy, factory, adapter, decorator, observer, builder, template method
   - Enterprise patterns: repository, unit of work, service layer, DTO, specification
4. **Domain-Driven Design**
   - Ubiquitous language, bounded contexts, context mapping
   - Entities, value objects, aggregates, domain events, domain services
5. **Development methodologies**
   - TDD, BDD, pair/mob programming
   - CQRS and event sourcing: introduction (deep dive in Module 19)
6. **Configuration & errors**
   - Twelve-Factor App; config precedence; feature toggles
   - Error handling strategy: domain errors vs infrastructure errors, wrapping, mapping to HTTP

**Labs:** refactor a "fat controller" into ports & adapters; event-storm the Gatepass domain.

**Gatepass M09:** reorganize into bounded contexts (Catalog, Inventory, Ordering, Ticketing, Identity) inside a modular monolith; enforce boundaries with lint rules.

---

## 10. Testing

**Goal:** ship with confidence and refactor without fear.

1. **Testing strategy**
   - Test pyramid vs trophy vs honeycomb; what to test where
2. **Unit tests**
   - Arrange/Act/Assert, table-driven tests, test doubles (stub, mock, fake, spy)
3. **Integration tests**
   - Real databases with Testcontainers; transactional test isolation; fixtures & factories
4. **Functional / end-to-end tests**
   - API-level black-box tests; test environments
5. **Contract testing**
   - Consumer-driven contracts (Pact); schema compatibility checks
6. **Advanced techniques**
   - Property-based testing, fuzzing, mutation testing, golden/snapshot tests
   - Concurrency/race testing, race detectors
7. **Quality practices**
   - Coverage (and its limits), flaky tests, test speed, TDD in practice
   - Performance/load testing: introduction (deep dive in Module 29)

**Labs:** property-test the seat-reservation logic; write a fuzz test for the HTTP parser from Module 05.

**Across languages:** `testing` + testify (Go), pytest/Hypothesis, Jest/Vitest, JUnit 5/Kotest/jqwik, xUnit/FsCheck, `cargo test`/proptest; Testcontainers exists for all.

**Gatepass M10:** full test suite; a concurrency test that fires 1,000 parallel reservations at one seat and asserts exactly one wins; CI-ready test commands.

---

## 11. Authentication & Authorization

**Goal:** know who the caller is and what they are allowed to do.

1. **Identity fundamentals**
   - Authentication vs authorization; identity, credentials, principals
2. **Passwords**
   - Hashing vs encryption; why MD5/SHA-1/SHA-256 alone are wrong
   - bcrypt, scrypt, Argon2; salts, peppers, work factors
   - Password policies, breached-password checks, reset flows
3. **Session-based auth**
   - Server-side sessions, cookie security, session fixation, logout & revocation
4. **Token-based auth**
   - Basic auth, bearer tokens, API keys
   - JWT: structure, signing algorithms (HS256/RS256/EdDSA), pitfalls (`alg: none`, no revocation)
   - Access + refresh tokens, rotation, token storage on clients
5. **Delegated & federated identity**
   - OAuth 2.0 / 2.1: roles, grant types, Authorization Code + PKCE, Client Credentials, Device flow
   - OpenID Connect: ID tokens, discovery, UserInfo
   - SAML and enterprise SSO; social login
6. **Strong authentication**
   - MFA, TOTP, WebAuthn/passkeys; account recovery
7. **Authorization models**
   - ACL, RBAC, ABAC, ReBAC (Zanzibar-style)
   - Policy engines (OPA, Cedar, OpenFGA); multi-tenancy and data isolation
8. **Service-to-service auth**
   - mTLS, signed requests (HMAC), workload identity

**Labs:** crack a leaked MD5 password table vs a bcrypt one; implement Authorization Code + PKCE against a local IdP (Keycloak); forge a JWT against a vulnerable verifier.

**Gatepass M11:** customer accounts (sessions), organizers & staff with RBAC, "Sign in with GitHub/Google" via OIDC, short-lived JWTs for the door-scanner app, API keys for organizer integrations.

---

## 12. Web Security

**Goal:** think like an attacker, build like a defender.

1. **Threat modeling**
   - STRIDE, attack surfaces, trust boundaries, data-flow diagrams
2. **OWASP Top 10, hands on**
   - Injection (SQL, command, LDAP), broken access control (IDOR), cryptographic failures
   - XSS (stored, reflected, DOM), CSRF, SSRF, insecure deserialization, XXE
   - Security misconfiguration, vulnerable components, logging failures
3. **Browser security model**
   - Same-origin policy, CORS (preflight, credentials), CSP, cookie attributes
   - Security headers: HSTS, `X-Content-Type-Options`, `Referrer-Policy`, `Permissions-Policy`
4. **Cryptography for engineers**
   - Hashing (SHA family, MD5 and why it's broken), MACs, symmetric (AES-GCM) and asymmetric (RSA, ECDSA, Ed25519) crypto
   - Encoding vs encryption vs hashing; key management; envelope encryption
5. **TLS in depth**
   - SSL history → TLS 1.2/1.3, handshake, cipher suites, PKI, certificate chains
   - Let's Encrypt/ACME, certificate rotation, mTLS
6. **Secrets & supply chain**
   - Secrets management (Vault, cloud KMS, SOPS); never in Git
   - Dependency scanning, SBOMs, signed artifacts, SAST/DAST
7. **Abuse prevention**
   - Rate limiting for security, bot protection, CAPTCHA trade-offs, account enumeration
8. **Privacy & compliance**
   - PII handling, data minimization, GDPR basics, PCI-DSS scope (why we don't store cards), audit logs

**Labs:** exploit and then fix SQLi, XSS, CSRF, IDOR and SSRF in a deliberately vulnerable Gatepass branch; configure a strict CSP.

**Gatepass M12:** threat model document; security headers & CORS policy; CSRF protection; login rate limiting; secrets moved to a secret manager; dependency scanning in CI; audit log for organizer actions.

---

# Part III: Data at Scale

## 13. Caching

**Goal:** make Gatepass fast without making it wrong.

1. **Fundamentals**
   - Latency numbers every engineer should know; hit ratio; working set
   - Where to cache: client, CDN, reverse proxy, application, database
2. **HTTP caching**
   - `Cache-Control` directives, `ETag`/`Last-Modified`, conditional requests, `Vary`
   - CDN caching, surrogate keys, purging
3. **Application caching patterns**
   - Cache-aside, read-through, write-through, write-behind, refresh-ahead
   - In-process vs distributed caches; multi-level caches
4. **Invalidation & consistency**
   - TTLs, explicit invalidation, versioned keys, event-driven invalidation
   - Stale-while-revalidate; consistency vs freshness trade-offs
5. **Failure modes**
   - Cache stampede/thundering herd (locking, request coalescing, probabilistic early expiry)
   - Hot keys, cache penetration, cold starts
6. **Eviction & internals**
   - LRU, LFU, ARC, TinyLFU; memory sizing
7. **Tools**
   - Redis, Memcached, Varnish; client-side caching

**Labs:** implement an LRU cache; reproduce a stampede and fix it with singleflight.

**Gatepass M13:** HTTP caching for public event pages (ETags, CDN-friendly headers), Redis cache-aside for event details, request coalescing for hot events, cache invalidation on event updates.

---

## 14. NoSQL & Beyond Relational

**Goal:** choose the right data store for each job.

1. **Theory**
   - CAP theorem, PACELC, BASE vs ACID
   - Data models and query patterns drive store choice
2. **Key-value stores**
   - Redis (data structures, TTLs, persistence, Lua, streams), DynamoDB
3. **Document databases**
   - MongoDB, Couchbase: schema design, embedding vs referencing, indexes
4. **Wide-column stores**
   - Cassandra, ScyllaDB, HBase: partition keys, clustering, query-first modeling
5. **Graph databases**
   - Neo4j, Neptune: nodes/edges, Cypher, when graphs win
6. **Time-series databases**
   - InfluxDB, TimescaleDB, Prometheus TSDB: retention, downsampling
7. **Column-oriented / OLAP**
   - ClickHouse, BigQuery, Redshift; OLTP vs OLAP; columnar storage
8. **Newer categories**
   - NewSQL (CockroachDB, Spanner, YugabyteDB), vector databases, embedded stores (SQLite, RocksDB)
9. **Polyglot persistence**
   - Keeping stores in sync; source of truth vs derived data

**Labs:** model the same data in Postgres, MongoDB and Cassandra and compare queries; build a leaderboard with Redis sorted sets.

**Gatepass M14:** seat holds in Redis with TTLs; "fans also bought" recommendations in a graph DB (or a Postgres graph query, with the trade-off written down); sales analytics in ClickHouse.

---

## 15. Search Engines

**Goal:** build search that users actually find useful.

1. **How search works**
   - Inverted indexes, tokenization, stemming, stop words, analyzers
   - Relevance: TF-IDF, BM25, boosting, scoring functions
2. **Engines**
   - Elasticsearch/OpenSearch, Solr, Meilisearch/Typesense; PostgreSQL full-text search
3. **Features**
   - Faceting & aggregations, filters vs queries, highlighting
   - Autocomplete (edge n-grams, completion suggesters), fuzzy matching, synonyms
   - Geo search, multi-language search
4. **Keeping the index fresh**
   - Dual writes vs CDC vs event-driven indexing; reindexing with aliases
5. **Operating search**
   - Shards, replicas, cluster sizing, relevance testing
6. **Semantic search**
   - Embeddings, vector search, hybrid ranking (intro, more in Module 31)

**Labs:** build a tiny inverted index from scratch; tune BM25 relevance with test queries.

**Gatepass M15:** event search with full text, facets (city, date, genre, price), geo radius, autocomplete; index fed by domain events; zero-downtime reindex.

---

## 16. File Storage & Media

**Goal:** handle files without killing your servers.

1. **Storage types**
   - Block, file and object storage; S3 semantics and consistency
2. **Uploads**
   - Multipart uploads, streaming vs buffering, size limits
   - Presigned URLs, direct-to-storage uploads, resumable uploads (tus)
   - Content validation, malware scanning
3. **Processing pipelines**
   - Image resizing/thumbnails, async processing, idempotent workers
4. **Delivery**
   - Signed URLs, CDNs for assets, range requests, streaming media basics
5. **Document generation**
   - PDFs, QR codes, signed/verifiable documents
6. **Lifecycle**
   - Retention, lifecycle rules, deletion & GDPR

**Labs:** stream a 5 GB upload with constant memory; build a thumbnail worker.

**Gatepass M16:** event poster upload via presigned URLs to MinIO/S3 with thumbnail generation; PDF tickets with signed QR codes delivered via signed download links.

---

## 17. Scaling Databases

**Goal:** grow past one database server.

1. **Scaling up first**
   - Vertical scaling, tuning, connection pooling, query optimization revisited
2. **Replication**
   - Leader-follower, multi-leader, leaderless (quorums)
   - Synchronous vs asynchronous; replication lag and read-your-writes
   - Failover, split brain
3. **Partitioning & sharding**
   - Range vs hash partitioning, consistent hashing, hot partitions
   - Table partitioning in Postgres; application-level sharding; rebalancing
4. **Distributed transactions**
   - Two-phase commit and its problems; sagas (deep dive in Module 19)
5. **Schema evolution at scale**
   - Zero-downtime migrations: expand/contract, backfills, online index creation
6. **Backups & recovery**
   - Logical vs physical backups, PITR, RPO/RTO, restore drills
7. **Database reliability**
   - Monitoring (slow queries, bloat, locks), capacity planning

**Labs:** set up streaming replication and observe lag; perform an expand/contract migration under load.

**Gatepass M17:** read replicas for browse traffic with read-your-writes for buyers; time-partitioned `orders` table; a zero-downtime column rename; a restore drill.

---

# Part IV: Asynchronous & Distributed Systems

## 18. Concurrency, Background Jobs & Reliability Patterns

**Goal:** do work outside the request cycle, and do it correctly.

1. **Concurrency hazards**
   - Race conditions, critical sections, mutexes, semaphores, atomics, deadlocks, livelocks
2. **Background processing**
   - Job queues (DB-backed, Redis-backed), workers, scheduling and cron
   - Priorities, concurrency limits, graceful worker shutdown
3. **Delivering reliably**
   - At-most-once, at-least-once, "exactly-once" (and why it's really effectively-once)
   - Idempotency, deduplication, idempotency keys
   - Retries with exponential backoff and jitter; poison messages
4. **Consistency between DB and side effects**
   - Dual-write problem; transactional outbox; inbox pattern
5. **Distributed locking**
   - Advisory locks, Redis locks (and the Redlock debate), fencing tokens, leases
6. **Notifications**
   - Email (SMTP, providers, deliverability), SMS, push; templating

**Across languages:** River/Asynq (Go), Celery/RQ/Dramatiq, BullMQ, Quartz/JobRunr, Hangfire, Sidekiq-style patterns everywhere.

**Labs:** reproduce a lost update with concurrent workers; implement an outbox relay.

**Gatepass M18:** hold-expiry worker, confirmation-email worker with retries, transactional outbox for all domain events, idempotent consumers.

---

## 19. Message Brokers & Event-Driven Architecture

**Goal:** decouple services with messages, without losing data.

1. **Messaging fundamentals**
   - Queues vs pub/sub vs logs/streams; push vs pull; acknowledgments
   - Ordering guarantees, partitioning, consumer groups
   - Delivery semantics; dead-letter queues; backpressure
2. **RabbitMQ**
   - Exchanges (direct, topic, fanout, headers), bindings, queues, prefetch, DLX
3. **Apache Kafka**
   - Topics, partitions, offsets, consumer groups, retention, compaction
   - Producers (acks, idempotence), transactions, Kafka Connect
4. **Other brokers**
   - NATS, Redis Streams, AWS SQS/SNS, Google Pub/Sub: when to use which
5. **Schemas & evolution**
   - Event contracts, schema registry, Avro/Protobuf, versioning events
6. **Event-driven patterns**
   - Event notification vs event-carried state transfer
   - Event sourcing: event store, projections, snapshots
   - CQRS: separate read/write models
   - Change Data Capture (Debezium)
7. **Sagas**
   - Choreography vs orchestration; compensating actions; workflow engines (Temporal)

**Labs:** compare RabbitMQ and Kafka under consumer failure; build a projection from an event log.

**Gatepass M19:** `OrderPlaced` → `PaymentCaptured` → `TicketsIssued` saga with compensation (release seats, refund); events on Kafka (or RabbitMQ); notification and analytics consumers.

---

## 20. Alternative API Styles

**Goal:** pick the right protocol for each boundary.

1. **GraphQL**
   - Schema, types, queries, mutations, subscriptions, resolvers
   - N+1 and DataLoader; pagination (Relay connections)
   - Security: depth/complexity limits, persisted queries; federation
2. **gRPC & Protocol Buffers**
   - IDL, code generation, unary and streaming RPCs
   - Deadlines, metadata, status codes, interceptors; gRPC-Web, gRPC gateway
3. **Webhooks**
   - Delivery, retries, signatures (HMAC), replay protection, receiving securely
4. **SOAP & XML-RPC**
   - WSDL, envelopes: enough to integrate with legacy systems
5. **Other styles**
   - JSON-RPC, tRPC, OData, AsyncAPI for event APIs
6. **Gateways & composition**
   - API gateways, Backend-for-Frontend (BFF), API aggregation

**Labs:** write a GraphQL query that brings a naive server down, then fix it; build a bidirectional gRPC stream.

**Gatepass M20:** internal gRPC between Ordering and a Payments service (with a fake payment provider); GraphQL API for the organizer dashboard; signed outgoing webhooks to organizers; verified incoming webhooks from the payment provider.

---

## 21. Real-Time Communication

**Goal:** push data to clients the moment it changes.

1. **Techniques**
   - Short polling, long polling, Server-Sent Events, WebSockets; comparison
2. **WebSockets in depth**
   - Upgrade handshake, frames, ping/pong, close codes, subprotocols
   - Authentication for long-lived connections
3. **Scaling real-time**
   - Connection limits, sticky sessions, fan-out with pub/sub (Redis, NATS)
   - Presence, backpressure, slow consumers, reconnect & resume
4. **Beyond**
   - WebRTC overview, WebTransport, MQTT for IoT

**Labs:** build a chat server with WebSockets; scale it to 2 nodes via Redis pub/sub.

**Gatepass M21:** live seat map over WebSockets (seats change colour as they're held/sold); SSE stream of queue position in the virtual waiting room.

---

## 22. Distributed Systems Fundamentals

**Goal:** understand why distributed systems are hard, and the tools that make them tractable.

1. **The fallacies of distributed computing**
2. **Failure models**
   - Crash, omission, Byzantine; partial failure; timeouts as the only signal
3. **Time & ordering**
   - Clock skew, NTP, logical clocks (Lamport), vector clocks, hybrid logical clocks, TrueTime
4. **Consistency models**
   - Linearizability, sequential, causal, eventual; read-your-writes, monotonic reads
5. **Consensus**
   - Raft (leader election, log replication), Paxos at a glance
   - etcd, ZooKeeper, Consul: what they're used for
6. **Coordination patterns**
   - Leader election, leases, fencing, membership & gossip
7. **Data structures for distribution**
   - Consistent hashing, CRDTs, Merkle trees, Bloom filters

**Labs:** run a Raft cluster (etcd) and kill nodes; demonstrate clock-skew bugs with simulated clocks.

**Gatepass M22:** leader election for the scheduler so only one instance runs hold expiry; fencing tokens on inventory updates; written analysis of Gatepass's consistency guarantees.

---

# Part V: Production Engineering

## 23. Containers, Virtualization & Orchestration

**Goal:** package and run Gatepass anywhere.

1. **Virtualization**
   - Hypervisors (type 1/2), VMs vs containers, microVMs (Firecracker)
2. **Container internals**
   - Namespaces, cgroups, union filesystems; LXC
3. **Docker**
   - Images, layers, Dockerfile best practices, multi-stage builds, distroless
   - Volumes, networks, Docker Compose for local development
   - Registries, image tagging, image scanning and signing
4. **Kubernetes**
   - Architecture: control plane, nodes, kubelet, etcd
   - Pods, Deployments, StatefulSets, Jobs/CronJobs, Services, Ingress
   - ConfigMaps, Secrets, resource requests/limits, probes
   - Autoscaling (HPA/VPA), rolling updates, PodDisruptionBudgets
   - Helm, Kustomize; operators (intro)

**Labs:** build a container by hand with `unshare` and `chroot`; shrink an image from 1 GB to 15 MB.

**Gatepass M23:** minimal images for all Gatepass components; `docker compose up` gives a full dev stack; Kubernetes manifests (kind/k3d) with probes and resource limits.

---

## 24. Web Servers, Proxies & Load Balancing

**Goal:** put the right thing in front of your application.

1. **Web servers**
   - Nginx, Apache, Caddy: architecture (process/event models) and configuration
   - Serving static files, compression, HTTP/2 & HTTP/3
2. **Proxies**
   - Forward vs reverse proxies; TLS termination; `X-Forwarded-*` and `Forwarded` headers
3. **Load balancing**
   - L4 vs L7; algorithms (round robin, least connections, consistent hashing, EWMA)
   - Health checks, connection draining, sticky sessions
   - HAProxy, Envoy, cloud load balancers
4. **Edge**
   - CDNs, edge caching, WAF, DDoS protection
5. **Service discovery**
   - DNS-based, registry-based (Consul), Kubernetes Services

**Labs:** load-balance three Gatepass instances with Nginx and kill one mid-load.

**Gatepass M24:** Caddy/Nginx in front of Gatepass with automatic TLS, compression, static assets, L7 routing to API/WebSocket backends, health-checked upstreams.

---

## 25. CI/CD & Infrastructure as Code

**Goal:** go from merge to production safely and repeatedly.

1. **Continuous Integration**
   - Pipelines: lint, test, build, scan; caching; matrix builds; test parallelism
   - GitHub Actions, GitLab CI, Jenkins, CircleCI
2. **Artifacts & releases**
   - Versioning, immutable artifacts, provenance, release notes
3. **Continuous Delivery/Deployment**
   - Environments (dev, staging, prod), promotion
   - Deployment strategies: rolling, blue-green, canary, shadow
   - Feature flags, dark launches, rollbacks
4. **Infrastructure as Code**
   - Terraform/OpenTofu, Pulumi; state, modules, plans; drift
   - Configuration management (Ansible) vs immutable infrastructure
5. **GitOps**
   - Argo CD, Flux; declarative deployments
6. **DevOps culture**
   - DORA metrics, trunk-based development, ownership

**Labs:** write a pipeline that blocks merges on failing tests; perform a canary rollout and automatic rollback.

**Gatepass M25:** GitHub Actions pipeline (lint, test, build, scan, push image); Terraform for infrastructure; canary deploy with automated rollback on error rate; feature flag for the waiting room.

---

## 26. Cloud Fundamentals

**Goal:** use the cloud deliberately and affordably.

1. **Service models**
   - IaaS, PaaS, SaaS, FaaS; shared responsibility model
2. **Core services (mapped across AWS / GCP / Azure)**
   - Compute: VMs, managed containers, serverless functions
   - Storage: object, block, file; managed databases and caches
   - Networking: VPCs, subnets, security groups, NAT, private endpoints
   - Messaging: queues, pub/sub, event buses
3. **Identity & access**
   - IAM users, roles, policies; least privilege; workload identity
4. **Reliability in the cloud**
   - Regions, availability zones, managed failover
5. **Cost**
   - Pricing models, egress, reserved/spot capacity, budgets and alerts, FinOps
6. **Serverless**
   - Cold starts, limits, event sources, when it fits and when it doesn't

**Gatepass M26:** deploy Gatepass to a cloud with managed Postgres, Redis, object storage and a managed broker, inside a private network, under a fixed monthly budget with alerts.

---

## 27. Observability

**Goal:** know what Gatepass is doing, and why, without guessing.

1. **Foundations**
   - Monitoring vs observability; telemetry types; instrumentation
2. **Logging**
   - Structured logging, levels, correlation IDs, sampling, PII redaction
   - Pipelines: Loki, ELK/OpenSearch, Fluent Bit
3. **Metrics**
   - Counters, gauges, histograms; cardinality
   - RED and USE methods; golden signals
   - Prometheus, PromQL, Grafana
4. **Distributed tracing**
   - Spans, context propagation (W3C Trace Context), sampling
   - OpenTelemetry SDKs & Collector; Jaeger/Tempo
5. **Profiling in production**
   - Continuous profiling (Pyroscope/Parca), eBPF tooling
6. **SLOs & alerting**
   - SLIs, SLOs, SLAs, error budgets, burn-rate alerts; alert fatigue
7. **Incident management**
   - On-call, runbooks, incident command, blameless postmortems

**Gatepass M27:** OpenTelemetry tracing across API → broker → workers; RED dashboards per service; SLO "99.9% of checkouts succeed within 2 s" with burn-rate alerts; runbooks.

---

## 28. Architectural Patterns

**Goal:** choose an architecture for the right reasons.

1. **Monoliths**
   - Monolith, modular monolith; when a monolith is the right answer
2. **Service-oriented architectures**
   - SOA, ESBs (historical context), microservices
   - Decomposition by business capability / bounded context; data ownership; database-per-service
   - Synchronous vs asynchronous communication; distributed monolith anti-pattern
3. **Serverless architecture**
   - FaaS, BaaS, event-driven serverless designs
4. **Service mesh**
   - Sidecars and ambient mesh (Istio, Linkerd); mTLS, retries, traffic shifting
5. **Other patterns**
   - Event-driven, space-based, cell-based architecture, micro-kernel/plugin
6. **Evolving architecture**
   - Strangler fig, branch by abstraction, anti-corruption layer
   - Conway's Law, team topologies, ADRs, fitness functions
7. **Twelve-Factor revisited**

**Gatepass M28:** extract Notifications and Search Indexer into separate services using the strangler-fig pattern; ADRs explaining what stays in the monolith and why.

---

## 29. Building for Scale & Resilience

**Goal:** stay up when traffic goes 100× and dependencies fail.

1. **Performance engineering**
   - Latency vs throughput, percentiles (p50/p99/p999), Little's Law, Amdahl's Law
   - Benchmarking, profiling under load, finding bottlenecks
2. **Scaling strategies**
   - Vertical vs horizontal scaling, statelessness, autoscaling, capacity planning
3. **Load testing**
   - k6, Gatling, Locust; load, stress, soak, spike tests; realistic traffic models
4. **Mitigation strategies**
   - Timeouts, retries with budgets, circuit breakers, bulkheads
   - Rate limiting algorithms: token bucket, leaky bucket, fixed/sliding window; distributed rate limiting
   - Throttling, backpressure, load shedding, load shifting, graceful degradation
   - Virtual waiting rooms/queues
5. **Resilience engineering**
   - Chaos engineering, fault injection, game days
6. **Migration strategies**
   - Big-bang vs incremental, parallel runs, dark traffic, data migration
7. **High availability & disaster recovery**
   - Redundancy, multi-AZ, multi-region (active-passive, active-active), RPO/RTO

**Labs:** find Gatepass's breaking point with k6; implement a token bucket in Redis; inject latency into the payment service.

**Gatepass M29:** flash-sale scenario: 50,000 virtual users at on-sale time. Virtual waiting room, rate limiting, circuit breaker on the payment provider, load shedding for non-critical endpoints, graceful degradation of search. Must hold the checkout SLO.

---

# Part VI: Synthesis

## 30. System Design

**Goal:** design systems on a whiteboard and defend the trade-offs.

1. **A framework**
   - Requirements (functional/non-functional), estimation (back-of-the-envelope), API, data model, high-level design, deep dives, bottlenecks
2. **Building blocks review**
   - Load balancers, caches, queues, databases, CDNs, search, object storage
3. **Case studies**
   - URL shortener, rate limiter, news feed, chat system, notification system
   - Distributed key-value store, web crawler, payment system, ride sharing
   - Ticket booking (compare with what you built)
4. **Writing design docs**
   - RFC/design doc structure, alternatives considered, review process

**Gatepass M30:** design doc for "Gatepass Global": multi-region, 10 million users, regional data residency, global on-sales.

---

## 31. AI-Powered Backends *(elective)*

**Goal:** integrate LLMs into a backend responsibly.

1. **LLM APIs**
   - Request/response, streaming (SSE), tokens, context windows, cost & latency
   - Structured output, tool/function calling
2. **Retrieval-Augmented Generation**
   - Embeddings, chunking, vector stores (pgvector), hybrid search, re-ranking
3. **Production concerns**
   - Caching, rate limits, fallbacks, timeouts; evaluation & regression tests
   - Prompt injection and data-leak risks; guardrails; PII
   - Observability for LLM calls
4. **Agents & MCP (intro)**
   - Exposing backend capabilities as tools safely

**Gatepass M31:** natural-language event search ("something chill in Lisbon on Friday under €30") via embeddings + filters; a support assistant answering from Gatepass's FAQ and order data with RAG, guarded against prompt injection.

---

## 32. Capstone: Launch Day

**Goal:** prove Gatepass is production-ready.

1. **Production readiness review**
   - Checklist: security, observability, scalability, backups, runbooks, on-call
2. **Game day**
   - Scripted failures during a simulated on-sale: DB failover, broker outage, payment provider down, bad deploy
3. **Postmortem**
   - Write a blameless postmortem for whatever broke
4. **Final presentation**
   - Architecture walkthrough, trade-offs, what you'd do differently

**Gatepass M32:** a fully deployed, documented, observable Gatepass that survives the game day, plus a portfolio-ready write-up.

---

## Appendix: Professional skills (threaded throughout)

- Reading code and documentation; asking good questions
- Writing: design docs, ADRs, PR descriptions, postmortems
- Estimation and breaking work down
- Code review etiquette
- Working with product, frontend, and ops counterparts
- Technical interviews: coding, system design, behavioural
