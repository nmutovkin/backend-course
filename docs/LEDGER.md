# Course Ledger

The course's memory. Read it before writing a module and update it after
(see [`AUTHORING.md`](./AUTHORING.md)). It must always describe what is
actually in the repository.

## 1. Status

| #  | Module                      | Status  | Notes |
|----|-----------------------------|---------|-------|
| 00 | Setup & Go Primer           | ✅ Done | Environment, tool installs, Go basics; no milestone code |
| 01 | How the Internet Works      | ✅ Done | `gpcurl` + `httpwire` (response side); self-contained labs with real captured output |
| 02 | Operating Systems & Linux   | ⏭ Next | |
| 03–32 | …                        | Planned | |

## 2. Canonical story & examples

Use these everywhere instead of inventing new examples.

| Item                    | Value                                                         | Introduced |
|-------------------------|---------------------------------------------------------------|------------|
| Product                 | Gatepass, event ticketing                                     | README     |
| Flagship event          | **The Midnight Owls at Riverside Hall, Lisbon**               | M01 §1     |
| Venue time zone         | `Europe/Lisbon` (to be used in M06)                           | M01 (implied) |
| Server region           | Frankfurt; fan in Lisbon; **RTT ≈ 40 ms**                     | M01 §1     |
| On-sale time            | 10:00:00 local time                                           | M01 §1     |
| Seat hold duration      | 10 minutes                                                    | PROJECT    |
| Flash-sale load         | 50,000 concurrent users                                       | PROJECT    |
| Example seats           | `A12`, `A13`                                                  | M01 §2.5   |
| Domains                 | `gatepass.test`, `api.gatepass.test` (`.test` is reserved, RFC 6761); `example.com` for real-world lookups | M01 |
| Example IPs             | server `203.0.113.7` / `2001:db8::7`; client `192.0.2.10`     | M01 §2.3   |
| Checkout SLO            | 99.9% of checkouts succeed within 2 s                          | PROJECT    |

## 3. Glossary: terms already defined

A term may be used freely after the module that defines it.

| Term(s)                                                        | Defined in |
|----------------------------------------------------------------|------------|
| Terminal basics (`pwd ls cd mkdir cat sudo`, `&`, Ctrl+C), `PATH`, `$` prompt convention | M00 §2.2–2.3 |
| Go: module/`go.mod`, package `main`, imports, stdlib, `:=`/`var`/`const`, zero values, basic types, explicit conversion, slices/`append`/`make`, maps, structs, exported vs unexported names | M00 §3.1–3.3 |
| Go: multiple results, `error`, sentinel errors, `fmt.Errorf` + `%w`, `errors.Is`, `_` | M00 §3.4 |
| Go: pointers, methods, value vs pointer receivers                | M00 §3.5   |
| Go: interfaces (implicit), `io.Reader`/`io.Writer`, `io.EOF`, wrapping readers, `bufio` | M00 §3.6 |
| Go: `defer`, packages & `internal/`, `cmd/<name>` layout         | M00 §3.7–3.8 |
| Go: `context.Context` (timeouts only; goroutines/channels only named) | M00 §3.9 |
| Go: tests, `t.Run`, `t.Errorf` vs `t.Fatalf`, table-driven tests, race detector (named), `go run/build/test/vet`, `gofmt` | M00 §3.10–3.11 |
| Big-endian / network byte order, DNS wire format (header flags, RCODEs, label encoding, type codes, RR layout, name compression), TCP DNS length prefix | M01 §3.2 |
| Reading `tcpdump` output (flags `S . P F R`, seq ranges), `ss`, `dig` output, `curl -w`, `tc netem` | M01 §6 |
| RTT, TTFB                                                      | M01 §1, §4.1 |
| TCP/IP & OSI layers, L4 vs L7, encapsulation                    | M01 §2.1   |
| IPv4/IPv6, CIDR, private ranges, loopback, link-local, NAT, BGP, traceroute/TTL, MAC/ARP | M01 §2.2 |
| Port, ephemeral port, 5-tuple, UDP, datagram, byte stream, **framing** | M01 §2.3 |
| Stub / recursive / authoritative resolver, DNS record types, DNS TTL, NXDOMAIN & negative caching, CNAME-at-apex, DNS load balancing, DoT/DoH, EDNS | M01 §2.4, §3.2 |
| HTTP request/response anatomy, status classes, HTTP versions    | M01 §2.5   |
| TLS goals, certificate (intro), SNI, ALPN, 0-RTT                | M01 §2.6, §3.4 |
| Same-origin policy (intro), CDN (intro), hosting types          | M01 §2.7   |
| Three-way handshake, sequence numbers, ACK, RTO, fast retransmit, flow control / `rwnd`, **backpressure** (TCP level), congestion control / `cwnd`, slow start, CUBIC/BBR, Nagle, delayed ACK, `TCP_NODELAY`, FIN/RST, TIME_WAIT, ephemeral port exhaustion, MSS | M01 §3.1 |
| Message-length rules, `Content-Length`, chunked encoding, trailers, request smuggling (named only) | M01 §3.3 |
| Keep-alive, pipelining, head-of-line blocking                  | M01 §3.3   |
| HTTP/2 frames, streams, HPACK; QUIC, HTTP/3, connection migration | M01 §3.4 |
| Happy Eyeballs                                                  | M01 §4.1   |
| Interoperability testing                                        | M01 §4.3   |

## 4. Open promises

Forward references made in published modules. Each must be fulfilled by the
target module, or moved here with a reason.

| Target | Promise                                                              | Made in |
|--------|----------------------------------------------------------------------|---------|
| 02     | Shell taught properly (M00 gave only the minimum); file descriptors as the real per-connection limit | M00 §2.2, M01 checkpoint 2 |
| 03     | Go in depth (`L3`): goroutines, channels, contexts properly, races   | M00 §3.9, §3.11 |
| 05     | Status codes in full, incl. redirect semantics (301/302/303/307/308)  | M01 §7   |
| 05     | Our server is strict: rejects HTTP/1.1 requests without `Host` (400)  | M01 Lab 2 |
| 06     | Why floats are wrong for money                                       | M00 §3.2 |
| 07     | Uses Docker (installed in M00 §2.5) to run PostgreSQL; explain just enough `docker run` | M00 §2.5 |
| 10     | Uses Docker for Testcontainers                                       | M00 §2.5 |
| 23     | Containers explained in depth                                        | M00 §2.5 |
| 05     | HTTP semantics: methods, status codes, headers, cookies               | M01 §2.5 |
| 05     | `httpwire` gains request parsing for our own server, with the same limits (8 KiB lines, 100 fields) | M01 §4, §4.2 |
| 07     | Connection pooling (DB); reuse rationale from M01 §1                  | M01 §1   |
| 10     | Interop testing generalized into contract testing                     | M01 §4.3 |
| 12     | TLS in depth: handshake internals, certificate chains, running TLS in production | M01 §2.6 |
| 12     | Cloud metadata IP `169.254.169.254` as an SSRF target                 | M01 §2.2 |
| 12     | Rate limiting by IP, with the NAT caveat (many users, one IP)         | M01 §2.2 |
| 12     | CORS as a consequence of the same-origin policy                       | M01 §2.7 |
| 12, 24 | HTTP request smuggling (TE vs CL disagreement)                        | M01 §3.3 |
| 12     | CAA records                                                           | M01 §2.4 |
| 13, 24 | CDN use for Gatepass                                                  | M01 §1, §2.7 |
| 18     | MX records / sending email                                            | M01 §2.4 |
| 19, 29 | Backpressure at the application level                                 | M01 §3.1 |
| 20     | gRPC runs on HTTP/2                                                    | M01 §3.4 |
| 24     | Keep-alive & HTTP/2 at the edge; HTTP/3 terminated by the edge        | M01 §1, §3.4 |
| 24     | DNS load balancing vs real load balancers; L4 vs L7                   | M01 §2.1, §2.4 |
| 24     | SRV records / service discovery                                       | M01 §2.4 |
| 26     | CIDR ranges in cloud networks (VPCs)                                  | M01 §2.2 |

## 5. Code state (`gatepass/`)

Go module `github.com/nmutovkin/backend-course/gatepass`, Go 1.24.

| Path                                  | Since | What it is |
|---------------------------------------|-------|------------|
| `internal/platform/httpwire`          | M01   | HTTP/1.1 wire format. `Header` (ordered `[]Field`, case-insensitive `Get`/`Values`), `Response`, `ReadResponse(r, method)`, internal `readLine`, `readHeader`, `chunkedReader`, `exactReader`. Limits `MaxLineLength` (8 KiB), `MaxHeaderCount` (100). Errors `ErrMalformed`, `ErrLineTooLong`, `ErrTooManyHeaders`. **No request side yet (M05).** |
| `cmd/gpcurl`                          | M01   | Raw-socket HTTP/1.1 client. `Client{Resolver, TLSConfig}`, `Fetch(ctx, Request, io.Writer) (*Response, *Timings, *Trace, error)`, `buildRequest`. Flags `-X -H -d -v -timeout`. Always `Connection: close`, first resolved address only. |

### Dependencies

| Module | Added in | Why |
|--------|----------|-----|
| *(none, stdlib only)* | | |

## 6. Decisions & conventions

| Decision                                                             | Made in |
|----------------------------------------------------------------------|---------|
| Go is the reference language; every module has *Across languages*    | README  |
| All project code in one Go module under `gatepass/`; layout follows `PROJECT.md` §4 | M01 |
| Shared infrastructure goes in `internal/platform/<name>`              | M01     |
| Stdlib only until a module justifies a dependency                     | M01     |
| HTTP headers are kept as an ordered list, not a map, at the wire level | M01 §4.2 |
| Parsers enforce size limits and reject ambiguous input                | M01 §4.2 |
| Errors are prefixed or wrapped with the phase or step that failed     | M01 §4.2 |
| CLI tools: data on stdout, diagnostics on stderr                      | M01 §4.2 |
| Lab solution code is not committed as files; short solutions/answers in `<details>`; extension tasks stay out of reference code | AUTHORING |
| **Modules are self-contained**: no external source is required; tools installed in M00 §2.3 (extend that table); labs show real output; every question answered | AUTHORING, after M01 |
| Lab conventions: local server `python3 -m http.server 8099` serving `/tmp/www`; `gpcurl` built to `/tmp/gpcurl`; scratch Go work in `~/goplay/<name>` | M01 §6 |
| Course OS baseline: Ubuntu 24.04 (native, WSL2, or container on macOS); Fedora/macOS equivalents given | M00 §2.1 |

## 7. Deviations from the plan

| Module | Deviation | Plan files updated? |
|--------|-----------|---------------------|
| 01     | 7 labs instead of the 4 listed (added `gpdig` DNS stretch lab, TIME_WAIT lab, netem RTT lab) | ✅ SYLLABUS |
| 01     | `gpcurl` lives in `cmd/gpcurl` (layout listed only api/worker/…) | ✅ PROJECT §4 |
| 00     | New Module 00 (Setup & Go Primer) added so Modules 01–02 don't depend on outside Go knowledge | ✅ SYLLABUS, README |
| 00     | Docker now required from Module 07 (README had said Part III) | ✅ README |

## 8. Notes for upcoming modules

- **M02:** M00 §2.2 taught only minimal shell; M02 §1 is the full treatment.
  `strace`, `lsof`, `ss` are already installed (M00 §2.3); add any new tool
  (e.g. `htop`) to that table. The lab server convention from M01 §6 can be reused.
- **M02:** the `gatepass` CLI (`cmd/gatepass`) is separate from `gpcurl`.
  Labs 1, 6 and 7 of M01 used `tcpdump`, `ss`, `tc`; M02 can build on them
  (`ss`, `lsof` appear in M02 §6). M01 §3.1 described TIME_WAIT; M02 §4 (file
  descriptors, `ulimit`) should link to it.
- **M03:** the domain model goes in `internal/<context>` packages per
  `PROJECT.md` §4 (e.g. `internal/catalog`, `internal/inventory`), even
  before M09 formalizes bounded contexts.
- **M05:** extend `httpwire` (add `Request`, `ReadRequest`, response
  writing) rather than starting a new package. Reuse `readLine`/`readHeader`
  and the limits. The fuzz-test lab in M10 targets this parser.
