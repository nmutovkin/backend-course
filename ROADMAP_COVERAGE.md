# roadmap.sh Backend Coverage

This file maps each topic on the [roadmap.sh backend roadmap](https://roadmap.sh/backend)
to the module(s) that cover it. Use it to check the course is complete.

| Roadmap topic                                   | Module(s)        |
|-------------------------------------------------|------------------|
| **Internet**: how it works, HTTP, DNS, domain names, hosting, browsers | 01 |
| **Basic frontend knowledge**: HTML, CSS, JavaScript | 05 |
| **OS & general knowledge**: terminal, processes, threads, memory, I/O, POSIX, networking | 02 |
| **Pick a language**: Go, Python, JS/TS, Java, C#, Rust, PHP, Ruby | 03 (+ *Across languages* in every module) |
| **Version control**: Git                         | 04               |
| **Repo hosting**: GitHub, GitLab, Bitbucket     | 04               |
| **Relational databases**: PostgreSQL, MySQL, MariaDB, MS SQL, Oracle, SQLite | 07 |
| **More about databases**: ORMs, ACID, transactions, N+1, normalization, failure modes, profiling performance | 07, 17 |
| **Database indexes**                            | 07               |
| **Data replication**                            | 17               |
| **Sharding strategies**                         | 17               |
| **CAP theorem**                                 | 14, 22           |
| **APIs**: REST, JSON APIs, SOAP, gRPC, GraphQL, HATEOAS, OpenAPI | 08, 20 |
| **Authentication**: cookie, basic, token, JWT, OAuth, OpenID, SAML | 11 |
| **Caching**: CDN, server-side, client-side, Redis, Memcached | 13 |
| **Web security**: HTTPS, CORS, CSP, OWASP risks, server security | 12 |
| **Hashing algorithms**: MD5, SHA, scrypt, bcrypt (+ Argon2) | 11, 12 |
| **SSL/TLS**                                     | 01, 12, 24       |
| **Testing**: unit, integration, functional      | 10               |
| **CI/CD**                                       | 25               |
| **Design & development principles**: GoF patterns, DDD, TDD, CQRS, event sourcing | 09, 19 |
| **Architectural patterns**: monolith, microservices, SOA, serverless, service mesh, twelve-factor | 09, 28 |
| **Search engines**: Elasticsearch, Solr         | 15               |
| **Message brokers**: RabbitMQ, Kafka            | 19               |
| **Containerization vs virtualization**: Docker, LXC, Kubernetes | 23 |
| **GraphQL**: Apollo, Relay                      | 20               |
| **NoSQL databases**: document, key-value, column, graph, time series | 14 |
| **Web servers**: Nginx, Apache, Caddy, MS IIS   | 24               |
| **Real-time data**: WebSockets, SSE, long polling, short polling | 21 |
| **Building for scale**: mitigation strategies (graceful degradation, throttling, backpressure, load shifting, circuit breaker), migration strategies, horizontal/vertical scaling | 29 |
| **Observability**: instrumentation, monitoring, telemetry | 27 |
| **Scaling databases**                           | 17               |

## Beyond the roadmap

Topics the course adds because real backend work needs them:

| Topic                                                       | Module |
|-------------------------------------------------------------|--------|
| Data formats, Unicode, time zones, money, identifiers       | 06     |
| File storage, uploads, media pipelines                      | 16     |
| Background jobs, outbox, idempotency, distributed locks     | 18     |
| Distributed systems theory: clocks, consensus, CRDTs        | 22     |
| Infrastructure as Code, GitOps                              | 25     |
| Cloud fundamentals & cost                                   | 26     |
| SLOs, incident management, postmortems                      | 27, 32 |
| System design interviews & design docs                      | 30     |
| LLM integration, RAG, AI safety in backends                 | 31     |
