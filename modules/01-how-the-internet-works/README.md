# Module 01: How the Internet Works

> **Depth:** Transport (TCP/UDP), DNS and HTTP at `L3`; networking models and
> addressing at `L2`; TLS, browsers and hosting at `L1` (TLS returns at `L2`
> in Module 12).
>
> **You will build:** `gpcurl`, an HTTP client on a raw TCP socket that
> shows where every millisecond of a request goes.
>
> **Time:** ~12 hours.

---

## 1. Why this exists

Gatepass's first big on-sale: **The Midnight Owls at Riverside Hall, Lisbon**.
Tickets go on sale at 10:00:00. A fan in Lisbon presses *Buy*, and our server
in Frankfurt has to answer.

Before our code runs a single line, the fan's laptop has to:

1. find out which IP address `gatepass.test` points to (**DNS**),
2. open a reliable connection to that address (**TCP**),
3. agree on encryption keys and check it is really talking to us (**TLS**),
4. send the request and wait for the answer (**HTTP**).

Lisbon to Frankfurt is roughly 40 ms there and back. That round trip is
called the **RTT** (round-trip time). Counting round trips:

| Step                         | Round trips | Time at 40 ms RTT |
|------------------------------|-------------|-------------------|
| DNS (resolver cache miss)    | 1+          | 40 ms or more     |
| TCP handshake                | 1           | 40 ms             |
| TLS 1.3 handshake            | 1           | 40 ms             |
| HTTP request → first byte    | 1 + server time | 40 ms + ours  |
| **Total before our code matters** |        | **~160 ms**       |

So a *new* connection costs about 160 ms, even if our server answers in zero
time. A *reused* connection costs 40 ms. At 50,000 fans per second that
difference decides whether the on-sale works, and it explains many decisions
later in the course: connection pooling (Module 07), keep-alive and HTTP/2
(Module 24), CDNs at the edge (Modules 13 and 24).

This module explains each of those steps until you could debug any of them
with a packet capture. Then you build a tool that measures them.

---

## 2. Concepts

### 2.1 Layers `L2`

Networking is built in layers. Each layer uses the one below it and hides its
details from the one above.

```
 TCP/IP model        What it does                       Unit       Examples
┌──────────────┐
│ Application  │  meaning of the conversation          message    HTTP, DNS, TLS*
├──────────────┤
│ Transport    │  process-to-process delivery          segment    TCP, UDP, QUIC*
├──────────────┤
│ Internet     │  host-to-host delivery across networks packet    IPv4, IPv6
├──────────────┤
│ Link         │  delivery to the next hop             frame      Ethernet, Wi-Fi
└──────────────┘
 * TLS sits between TCP and HTTP; QUIC is a transport built on top of UDP.
```

The OSI model splits the same stack into seven layers (adding *physical*,
*session* and *presentation*). People still say "L4 load balancer" (transport)
and "L7 load balancer" (application) using OSI numbers, so learn the numbers
even though TCP/IP is what's actually implemented.

**Encapsulation:** when you send an HTTP request, each layer wraps the data
from the layer above in its own header:

```
[ Ethernet hdr | IP hdr | TCP hdr | GET /events HTTP/1.1 ... | Ethernet trailer ]
                                   └──── HTTP message ───────┘
                          └──────── TCP segment ─────────────┘
                 └──────────────── IP packet ────────────────┘
└──────────────────────────── Ethernet frame ──────────────────────────────────┘
```

Routers along the way only look at the IP header. Only the two endpoints look
at TCP and HTTP. (Proxies and load balancers are deliberate exceptions, as
Module 24 will show.)

### 2.2 Addressing and routing `L2`

**IPv4** addresses are 32 bits, written `203.0.113.7`. **IPv6** addresses are
128 bits, written `2001:db8::7`. There are only about 4.3 billion IPv4
addresses, so most devices share one through NAT (below).

**CIDR** notation describes a range: `10.0.0.0/16` means "the first 16 bits are
fixed", so 65,536 addresses from `10.0.0.0` to `10.0.255.255`. You will write
CIDR ranges constantly when configuring cloud networks (Module 26) and
firewalls.

Ranges worth knowing:

| Range                                        | Meaning                                 |
|----------------------------------------------|-----------------------------------------|
| `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16` | Private (RFC 1918): not routed on the Internet |
| `127.0.0.0/8`, `::1`                         | Loopback: this machine                  |
| `169.254.0.0/16`                             | Link-local; also the cloud metadata service `169.254.169.254` (an SSRF target, Module 12) |
| `0.0.0.0`                                    | "All interfaces" when binding a server  |
| `192.0.2.0/24`, `198.51.100.0/24`, `203.0.113.0/24`, `2001:db8::/32` | Reserved for documentation (used in this course) |

**NAT** (network address translation) lets a home router give every device a
private address while sharing one public address. The router rewrites
outgoing packets and remembers the mapping so replies find their way back.
For backends this means that *many users can arrive from the same IP address*
(a whole office, a mobile carrier). Remember this when we rate-limit by IP in
Module 12.

**Routing:** each router forwards a packet to the next hop based on the
destination address and its routing table. Between big networks, routes are
exchanged with BGP `L1`. `traceroute` shows the hops by sending packets with
an increasing TTL (time-to-live, a hop counter). Each router that decrements
the TTL to zero replies with an error, revealing itself.

On a local network, IP addresses are mapped to hardware (**MAC**) addresses
using **ARP** (IPv4) or Neighbor Discovery (IPv6).

### 2.3 Transport: TCP and UDP `L3`

IP delivers packets *between machines*, with no guarantees: packets can be
lost, duplicated, reordered or corrupted. The transport layer delivers data
*between processes* and decides what guarantees to add.

A process is identified by a **port** (0–65535). A connection is identified
by its **5-tuple**:

```
(protocol, source IP, source port, destination IP, destination port)
(TCP,      192.0.2.10, 51514,      203.0.113.7,     443)
```

The server's port is fixed and well known (443 for HTTPS). The client's port
is an **ephemeral port** that the OS picks from a range (on Linux,
`/proc/sys/net/ipv4/ip_local_port_range`, by default 32768–60999). The
5-tuple is why one server port can hold millions of connections: each one
differs in the client's IP or port.

**UDP** adds ports and a checksum to IP, and nothing else. No connection, no
retransmission, no ordering. Each `send` is one datagram. It's the right choice
when a late answer is worthless (live video, games), when one request and one
response fit in a packet (DNS), or when you want to build your own reliability
on top (QUIC).

**TCP** gives you a reliable, ordered **byte stream**. Section 3 shows how. One
consequence matters for every protocol on top of TCP: **TCP has no idea where
your messages begin and end.** If you write `"hello"` then `"world"`, the
other side may read `"hel"`, `"loworld"`, or `"helloworld"` in a single read.
Every protocol on TCP therefore needs its own **framing** rules to mark message
boundaries. HTTP/1.1's framing rules are what `gpcurl` implements in
section 4.

### 2.4 DNS `L3`

DNS maps names to records. It's a distributed, hierarchical, heavily cached
database.

```
                       . (root)
          ┌────────────┼──────────────┐
         com          org            test
          │                            │
       example                     gatepass
          │                            │
         www                          api
```

Roles in a lookup:

- **Stub resolver:** a library in your OS or language runtime. It asks one
  server and trusts the answer.
- **Recursive resolver:** your ISP's resolver, or `1.1.1.1` or `8.8.8.8`. It
  does the actual work and caches the results.
- **Authoritative servers:** they hold the real records for a zone. The root
  servers know who is authoritative for `.com`, the `.com` servers know who is
  authoritative for `example.com`, and so on.

Record types you will use:

| Type    | Holds                                    | Example use                         |
|---------|------------------------------------------|-------------------------------------|
| `A`     | IPv4 address                             | `api.gatepass.test → 203.0.113.7`   |
| `AAAA`  | IPv6 address                             | `api.gatepass.test → 2001:db8::7`   |
| `CNAME` | Alias to another name                    | `www → gatepass-prod.cdn.example`   |
| `MX`    | Mail servers, with priority              | Sending email (Module 18)           |
| `TXT`   | Arbitrary text                           | SPF/DKIM for email, domain verification |
| `NS`    | Authoritative name servers for a zone    | Delegation                          |
| `SRV`   | Host and port for a service              | Service discovery (Module 24)       |
| `CAA`   | Which CAs may issue certificates         | TLS hygiene (Module 12)             |

Every record has a **TTL** in seconds: how long others may cache it. This is
the truth behind "DNS propagation". Nothing propagates. Caches all over the
world hold the old answer until its TTL runs out. If you plan to move
Gatepass to new servers, **lower the TTL a day in advance** (say from 3600 to
60). Afterwards you can raise it again.

Two gotchas:

- **Negative answers are cached too.** If you look up a name before it exists,
  the "no such name" (NXDOMAIN) answer gets cached for the zone's negative TTL.
- **A `CNAME` can't sit at the zone apex** (`gatepass.test` itself), because
  the apex must also have `NS` and `SOA` records, and a CNAME can't coexist
  with other records. DNS providers work around this with "ALIAS" or
  "flattening" records.

DNS is also a crude load balancer: return several `A` records and clients
spread across them. But clients cache answers and ignore the TTL more often
than you would like, so failover through DNS is slow. Module 24 compares this
with real load balancers.

### 2.5 HTTP `L3`

HTTP is a request/response protocol. The client sends a request:

```
POST /orders HTTP/1.1                       ← request line: method, target, version
Host: api.gatepass.test                      ← header fields (name: value)
Content-Type: application/json
Content-Length: 41
                                             ← empty line: end of headers
{"event_id":"…","seats":["A12","A13"]}      ← body (exactly 41 bytes)
```

and the server sends a response:

```
HTTP/1.1 201 Created                         ← status line: version, code, reason
Content-Type: application/json
Content-Length: 27

{"order_id":"…","total":…}
```

Status codes come in classes: `1xx` informational, `2xx` success, `3xx`
redirection, `4xx` the client's mistake, `5xx` the server's mistake. The
semantics of methods, status codes and headers are Module 05's job. This
module cares about how the bytes travel.

HTTP versions, briefly. Section 3.4 goes deeper.

| Version  | Year | Key idea                                                              |
|----------|------|-----------------------------------------------------------------------|
| HTTP/1.0 | 1996 | One request per TCP connection                                        |
| HTTP/1.1 | 1997 | Persistent connections (keep-alive), `Host` header, chunked bodies    |
| HTTP/2   | 2015 | Binary frames, many concurrent streams over one TCP connection        |
| HTTP/3   | 2022 | HTTP/2's ideas over QUIC (UDP), so one lost packet doesn't stall every stream |

### 2.6 TLS and HTTPS `L1`

HTTPS is HTTP inside a **TLS** session. TLS gives you three things:

- **Confidentiality:** nobody in between can read the traffic.
- **Integrity:** nobody can modify it undetected.
- **Authentication:** the server proves it owns `gatepass.test` with a
  **certificate** signed by a certificate authority the client already trusts.

The TLS 1.3 handshake costs one round trip after TCP (TLS 1.2 cost two). The
client says which hostname it wants (**SNI**, Server Name Indication) and
which application protocols it speaks (**ALPN**, e.g. `h2` or `http/1.1`).
`gpcurl` sets both. How the handshake really works, certificate chains, and
running TLS in production are covered in Module 12.

### 2.7 Browsers, hosting and CDNs `L1`

When you type a URL, a browser does everything above, then parses HTML and
fetches dozens more resources: CSS, JavaScript, images, API calls. It also
enforces the **same-origin policy**: a page from one origin (scheme + host +
port) can't freely read responses from another. That policy is why CORS
exists (Module 12).

Where Gatepass can live, from most to least control:

- **Bare metal:** your own hardware.
- **VPS or cloud VM:** a virtual machine you manage.
- **Containers or PaaS:** you ship an image or code; the platform runs it.
- **Serverless:** you ship functions; the platform runs them per request.

We start on a laptop and reach managed cloud in Module 26.

A **CDN** (content delivery network) is a fleet of servers close to users that
cache and serve content, so a fan in Lisbon gets the event page from Lisbon
instead of Frankfurt, cutting the 40 ms RTT to a few milliseconds. CDNs are
also where TLS handshakes are cheapest (short RTT). Modules 13 and 24 use one.

---

## 3. Under the hood

### 3.1 The TCP connection lifecycle `L3`

```
 Client (192.0.2.10:51514)                       Server (203.0.113.7:443)
        │                                                  │  LISTEN
        │ ── SYN  seq=x ─────────────────────────────────▶ │
SYN_SENT│                                                  │  SYN_RECEIVED
        │ ◀──────────────────── SYN+ACK  seq=y, ack=x+1 ── │
        │ ── ACK  ack=y+1 ───────────────────────────────▶ │
ESTABLISHED                                                │  ESTABLISHED
        │                                                  │
        │ ── data (request) ─────────────────────────────▶ │
        │ ◀─────────────────────────────── data (response) │
        │                                                  │
        │ ── FIN ────────────────────────────────────────▶ │  CLOSE_WAIT
FIN_WAIT│ ◀───────────────────────────────────────── ACK ──│
        │ ◀───────────────────────────────────────── FIN ──│  LAST_ACK
        │ ── ACK ────────────────────────────────────────▶ │  CLOSED
TIME_WAIT  (waits 2×MSL, 60 s on Linux)
        │
     CLOSED
```

**The handshake** exchanges initial **sequence numbers** (x and y, chosen
randomly). From then on, every byte in each direction is numbered, so the
receiver can put segments in order, drop duplicates and spot gaps. The
receiver sends **ACKs** ("I have everything up to byte N"). If the sender sees
no ACK before its retransmission timeout (**RTO**, derived from measured RTT),
it resends. Three duplicate ACKs trigger a *fast retransmit* without waiting.

The handshake costs one RTT before any data can flow. That's why connection
reuse matters so much (section 1).

**Flow control** protects the *receiver*. Each ACK carries a receive window
(`rwnd`): "I have room for N more bytes." The sender never has more than
`rwnd` unacknowledged bytes in flight. If your server reads slowly from a
socket, the window shrinks to zero and the sender stalls. That's
*backpressure*, built into TCP. We use the same idea at the application level
in Modules 19 and 29.

**Congestion control** protects the *network*. The sender keeps a congestion
window (`cwnd`) and sends at most `min(cwnd, rwnd)` bytes per RTT:

- **Slow start:** a new connection starts with a small `cwnd` (10 segments,
  about 14 KB, on modern Linux) and doubles it every RTT.
- **On loss:** `cwnd` is cut (halved for Reno, less for CUBIC). BBR, used by
  Google, instead models bandwidth and RTT directly.

The practical result: **a fresh connection is slow.** It needs several RTTs
to reach full speed, so a 1 MB response over a new connection takes several
round trips even on a fast link. Another reason to reuse connections, and one
reason to keep event-page responses small.

**Nagle's algorithm** holds back small writes until earlier data is
acknowledged, to avoid sending many tiny packets. Combined with the
receiver's *delayed ACK*, it can add up to ~40 ms to request/response
protocols. Most servers and clients turn it off with `TCP_NODELAY`. Go does
this by default for every TCP connection.

**Closing.** Either side sends `FIN` ("I'm done sending"). Each direction
closes separately, so TCP is *full-duplex*. A **RST** aborts the connection
immediately. You see it as "connection reset by peer", typically when a
process crashes or a proxy gives up on a connection.

**TIME_WAIT.** The side that closes first waits about 60 s before forgetting
the connection, so that late packets from the old connection can't be
mistaken for a new one with the same 5-tuple. A client that opens and closes
thousands of connections per second to the same server ends up with tens of
thousands of sockets in `TIME_WAIT`. Eventually it runs out of ephemeral
ports and new connections fail with `EADDRNOTAVAIL` ("cannot assign
requested address"). This happens in real life when a service creates a new
HTTP client per request instead of sharing a connection pool. Lab 6 makes it
happen.

**TCP header, the fields that matter:**

| Field                  | Size     | Purpose                                              |
|------------------------|----------|------------------------------------------------------|
| Source / dest port     | 16 + 16 bits | Which processes are talking                      |
| Sequence number        | 32 bits  | Number of the first data byte in this segment        |
| Acknowledgment number  | 32 bits  | Next byte expected from the other side               |
| Flags                  | 1 bit each | `SYN`, `ACK`, `FIN`, `RST`, `PSH`, `URG`, `ECE`, `CWR` |
| Window                 | 16 bits  | Receive window `rwnd` (scaled by the window-scale option) |
| Checksum               | 16 bits  | Detects corruption                                   |
| Options                | variable | MSS, window scale, SACK, timestamps                  |

The **MSS** option
(maximum segment size, typically 1460 bytes on Ethernet) caps how much data
fits in one segment.

### 3.2 A DNS lookup, step by step `L3`

Resolving `api.gatepass.test` with empty caches:

```
 Stub (your app)      Recursive resolver          Authoritative servers
      │  A api.gatepass.test?   │
      │ ──────────────────────▶ │  ── A api.gatepass.test? ──▶  root (.)
      │                         │  ◀─ "ask the .test servers" (NS + glue)
      │                         │  ── A api.gatepass.test? ──▶  .test TLD
      │                         │  ◀─ "ask ns1.gatepass.test"   (NS + glue)
      │                         │  ── A api.gatepass.test? ──▶  ns1.gatepass.test
      │                         │  ◀─ A 203.0.113.7, TTL 300
      │ ◀── A 203.0.113.7 ───── │  (caches every answer for its TTL)
```

In practice the root and TLD answers are almost always cached, so a lookup
is usually one round trip to the recursive resolver, or no round trip at all
if the OS cache has the answer.

Caches stack up: inside the application runtime (some runtimes cache; Go's
does not), in the OS (`systemd-resolved`, `nscd`), in the recursive
resolver, and in the browser. When you debug "DNS is returning the old IP",
work out which cache is answering.

**On the wire.** A DNS message is a compact binary format, usually in a
single UDP datagram to port 53:

```
┌──────────────────────── Header (12 bytes) ────────────────────────┐
│ ID (16 bits)   │ flags: QR, opcode, AA, TC, RD, RA, rcode          │
│ QDCOUNT        │ ANCOUNT        │ NSCOUNT        │ ARCOUNT          │
├──────────────────────── Question ─────────────────────────────────┤
│ QNAME: 3 'a' 'p' 'i' 8 'g' 'a' 't' 'e' 'p' 'a' 's' 's' 4 't' 'e'   │
│        's' 't' 0          (length-prefixed labels, then a 0 byte)  │
│ QTYPE: 1 (A)   │ QCLASS: 1 (IN)                                    │
├──────────────────────── Answer(s) ────────────────────────────────┤
│ NAME (often a 2-byte pointer back to QNAME: 0xC00C)               │
│ TYPE │ CLASS │ TTL (32 bits) │ RDLENGTH │ RDATA (4 bytes for an A) │
└───────────────────────────────────────────────────────────────────┘
```

- The client picks a random **ID** and matches the reply by it. Randomizing
  the ID and the source port is what defends against cache-poisoning attacks.
- `RD=1` means "recursion desired".
- If the answer doesn't fit in the datagram, the server sets **TC**
  (truncated) and the client retries over TCP. **EDNS(0)** lets clients
  advertise larger UDP buffers.
- **DoT** and **DoH** (DNS over TLS / over HTTPS) encrypt this traffic.

Lab 5 has you build this packet by hand.

### 3.3 HTTP/1.1 framing: where does a message end? `L3`

This is the heart of `gpcurl`, and the source of real security bugs. Because
TCP is just a stream of bytes (section 2.3), a reader needs rules to know
where a response body ends. RFC 9112 §6.3 gives them, in priority order:

1. **No body, whatever the headers say:** responses to `HEAD`, and status
   codes `1xx`, `204` and `304`.
2. **`Transfer-Encoding: chunked`:** the body is sent as a series of chunks,
   each prefixed with its size in hex, ending with a zero-size chunk:
   ```
   4\r\n
   Wiki\r\n
   7\r\n
   pedia i\r\n
   0\r\n
   \r\n
   ```
   The server can stream a response without knowing its length in advance.
   `Transfer-Encoding` *wins over* `Content-Length` if both are present.
3. **`Content-Length: N`:** read exactly N bytes. If several values disagree,
   the message is invalid.
4. **Otherwise:** the body runs until the server closes the connection
   (an HTTP/1.0 habit; then you can't tell a complete response from a
   truncated one).

Why so strict? If a proxy and the server behind it disagree about where a
message ends, an attacker can hide a second request inside the body of the
first. That's **HTTP request smuggling**, covered in Modules 12 and 24. Our
parser takes the conservative path and rejects ambiguous input: conflicting
`Content-Length` values, obsolete line folding, overlong lines.

**Keep-alive and its limit.** HTTP/1.1 connections are persistent by default:
after one response, the next request can reuse the connection
(`Connection: close` opts out, as `gpcurl` does to stay simple). But only one
request can be in flight per connection at a time. *Pipelining*, sending
several requests without waiting, is in the spec but broken in practice,
because responses must come back in order and one slow response blocks all
the others: **head-of-line (HOL) blocking**. Browsers work around it by
opening ~6 connections per host.

### 3.4 HTTP/2 and HTTP/3 `L3`

**HTTP/2** keeps HTTP's semantics (methods, headers, status codes) but changes
the wire format completely:

- Everything is a binary **frame** with a 9-byte header (length, type, flags,
  stream ID). Common types: `HEADERS`, `DATA`, `SETTINGS`, `WINDOW_UPDATE`,
  `RST_STREAM`, `GOAWAY`.
- Each request/response pair is a **stream** with its own ID. Frames from
  many streams interleave on **one TCP connection**, which removes HTTP-level
  HOL blocking.
- Headers are compressed with **HPACK**: a table of previously sent headers
  is shared, so repeated headers like cookies cost a few bytes.
- Per-stream flow control, like TCP's but one level up.

The catch: everything still rides on one TCP connection. **One lost packet
stalls every stream** until it's retransmitted, because TCP only delivers
bytes in order. That's HOL blocking moved down to the transport layer. On
lossy mobile networks, HTTP/2 can be slower than several HTTP/1.1
connections.

**HTTP/3** fixes that by replacing TCP with **QUIC**, a transport built on UDP:

- Streams are independent at the transport level: a lost packet only stalls
  its own stream.
- TLS 1.3 is built in, so connection + encryption takes **1 RTT**, and
  repeat visits can send data in the first packet (**0-RTT**, with replay
  caveats).
- Connections are identified by a connection ID, not the 5-tuple, so a phone
  moving from Wi-Fi to 4G keeps its connection.

**What a backend actually speaks.** Typically browsers talk HTTP/2 or HTTP/3
to the CDN or load balancer. That edge talks HTTP/1.1 or HTTP/2 to your
application over a fast private network, where packet loss is rare and the
RTT is under a millisecond. Your app rarely needs HTTP/3 itself; the edge
handles it (Module 24). Internal services often use HTTP/2 because gRPC
requires it (Module 20).

---

## 4. Build it: `gpcurl`

**Goal:** an HTTP/1.1 client written directly on a TCP socket, with its own
response parser, that prints the duration of every phase from section 1.

**Code:**

- [`gatepass/internal/platform/httpwire`](../../gatepass/internal/platform/httpwire/httpwire.go):
  the HTTP/1.1 wire format. Module 05 extends this package with request
  parsing for our own server.
- [`gatepass/cmd/gpcurl`](../../gatepass/cmd/gpcurl/): the client.

Run it:

```console
$ cd gatepass
$ go run ./cmd/gpcurl -v http://localhost:8099/ > /dev/null
* resolved 127.0.0.1
* connected 127.0.0.1:37580 -> 127.0.0.1:8099
> GET / HTTP/1.1
> Host: localhost:8099
> User-Agent: gpcurl/0.1
> Accept: */*
> Connection: close
< HTTP/1.0 200 OK
< Server: SimpleHTTP/0.6 Python/3.13.16
< Content-type: text/html; charset=utf-8
< Content-Length: 224
timings:
  DNS lookup          177µs
  TCP connect         525µs
  Server (TTFB)     3.658ms
  Content              50µs
  Total              4.41ms
```

(Any local server works: `python3 -m http.server 8099` is the quickest.) Try
it against a real HTTPS site to see the TLS phase and real RTTs.

Notice `127.0.0.1:37580` in the `connected` line: that's the ephemeral port
from section 2.3.

### 4.1 Walking through `Fetch`

[`fetch.go`](../../gatepass/cmd/gpcurl/fetch.go) does the five phases in
order and timestamps each one.

**Phase 1: DNS.** If the host is already an IP address, skip the lookup;
otherwise ask the resolver:

```go
addrs, err := c.Resolver.LookupNetIP(ctx, "ip", host)
```

Go has two resolvers: a pure-Go one that reads `/etc/resolv.conf` and speaks
DNS itself, and one that calls the C library (`getaddrinfo`). Which one runs
depends on the platform and the `GODEBUG=netdns=go|cgo` setting. Other
languages almost always go through the C library.

**Phase 2: TCP connect.** `DialContext` calls the `socket` and `connect`
system calls. The three-way handshake from section 3.1 happens inside the
kernel, and `DialContext` returns once it completes. We take the first
address. Real clients use *Happy Eyeballs* (RFC 8305): they try IPv6 and IPv4
in parallel and keep whichever connects first.

**Phase 3: TLS** (https only). We wrap the TCP connection in a TLS client.
`ServerName` sets SNI and is the name the certificate must match.
`NextProtos` advertises ALPN `http/1.1`, so the server won't try to speak
HTTP/2 to us.

**Phase 4: request and first byte.** `buildRequest` writes the request by
hand: request line, `Host` (the only mandatory header in HTTP/1.1, and how
one IP address serves many sites), our headers, `Connection: close`, an empty
line, then the body. We wrap the connection in a `firstByteReader` that
records the moment the first response byte arrives. **TTFB** (time to first
byte) is network RTT plus server processing time. It's the number you
optimize from Module 07 on.

**Phase 5: content.** `httpwire.ReadResponse` parses the status line and
headers, then picks a body reader using exactly the four rules of
section 3.3:

```go
if method == "HEAD" || code < 200 || code == 204 || code == 304 { /* no body */ }
if te := resp.Header.Get("Transfer-Encoding"); te != "" { /* chunked, or until close */ }
if cls := resp.Header.Values("Content-Length"); len(cls) > 0 { /* exactly N bytes */ }
return r, nil // until the connection closes
```

`chunkedReader` is a small state machine: read a hex size line, hand out
that many bytes, expect `\r\n`, repeat until a zero-size chunk, then read
the trailer fields.

### 4.2 Design choices worth noticing

- **Headers are an ordered list, not a map.** On the wire, field order and
  duplicates exist (`Set-Cookie` appears many times). We keep the message as
  it arrived and offer case-insensitive lookups on top.
- **Limits everywhere.** 8 KiB per line, 100 header fields. A peer that sends
  an endless line otherwise eats your memory. Module 05 applies the same rule
  to our server.
- **Errors say which phase failed** (`dns:`, `connect:`, `tls:`, `read:`).
  When something breaks in production, the first question is always "which
  layer?"
- **Body on stdout, everything else on stderr,** so `gpcurl URL > file` works.
- **Stdlib only.** No dependencies until a module gives a reason for one.

### 4.3 Tests

```console
$ go test -race ./...
```

- `httpwire_test.go` checks every framing rule plus the malformed cases:
  conflicting lengths, folded headers, bad chunk sizes, truncated bodies.
- `fetch_test.go` runs `gpcurl` against real `net/http` test servers over
  plain HTTP and HTTPS. It checks a chunked response decodes correctly, a
  request body arrives, the `Host` header is right, and an untrusted
  certificate is rejected.

Testing against the standard library's server is a cheap form of
**interoperability testing**: our code is checked against an independent
implementation of the same RFC. Module 10 generalizes this into contract
testing.

---

## 5. Across languages

Every language wraps the same operating-system socket API (BSD sockets:
`socket`, `connect`, `read`, `write`, `close`). Only the names differ.

| Language    | Raw TCP                          | DNS lookup                         | TLS client                       |
|-------------|----------------------------------|------------------------------------|----------------------------------|
| Go          | `net.Dial("tcp", addr)`          | `net.Resolver.LookupNetIP`         | `crypto/tls.Client`              |
| Python      | `socket.create_connection`       | `socket.getaddrinfo`               | `ssl.create_default_context().wrap_socket` |
| Node / TS   | `net.connect`                    | `dns.promises.lookup`              | `tls.connect`                    |
| Java/Kotlin | `java.net.Socket` / NIO `SocketChannel` | `InetAddress.getAllByName`  | `SSLSocketFactory`               |
| C#          | `System.Net.Sockets.TcpClient`   | `Dns.GetHostAddressesAsync`        | `SslStream`                      |
| Rust        | `std::net::TcpStream` / `tokio::net` | `ToSocketAddrs` / `tokio::net::lookup_host` | `rustls` / `native-tls`   |

Everyday HTTP clients (`net/http`, `requests`/`httpx`, `fetch`/`undici`,
`java.net.http.HttpClient`/OkHttp, `HttpClient`, `reqwest`) do everything
`gpcurl` does, plus connection pooling, redirects, compression, proxies and
HTTP/2. **The one rule that transfers to all of them:** create the client
once and reuse it, so its connection pool can reuse TCP+TLS connections
(section 3.1, Lab 6).

---

## 6. Labs

1. **Watch the handshake.** Start a local server and capture traffic:
   `sudo tcpdump -i lo -n port 8099` (or Wireshark on the loopback
   interface). Run `gpcurl` against it. Identify the SYN, SYN-ACK and ACK, the
   request segment, the response, and the FIN/ACK exchange. Which side sends
   the first FIN? Why? (Hint: the `Connection: close` header.)

2. **Speak HTTP by hand.** Run `nc localhost 8099`, type
   `GET / HTTP/1.1`, `Host: localhost` and an empty line. Now try:
   no `Host` header; `HTTP/1.0` instead of `1.1`; an invalid method. Note how
   the server responds to each.

3. **Follow the DNS chain.** On a machine with internet access, run
   `dig +trace example.com` and identify the root, TLD and authoritative
   servers. Then run `dig example.com` twice against your resolver and
   compare the TTLs. What do you notice?

4. **Compare with curl.** Run
   `curl -o /dev/null -s -w 'dns %{time_namelookup} connect %{time_connect} tls %{time_appconnect} ttfb %{time_starttransfer} total %{time_total}\n' https://example.com`
   next to `gpcurl https://example.com > /dev/null`. curl's numbers are
   cumulative; ours are per-phase. Convert one into the other and check they
   roughly agree. Which phase is closest to one RTT? Check with `ping`.

5. **Stretch, `L3` DNS: build `gpdig`.** Write a tool that builds the DNS query
   from section 3.2 byte by byte, sends it over UDP to a resolver (`8.8.8.8:53`
   or the one in `/etc/resolv.conf`), and prints the `A` records and TTLs
   from the answer. Handle name compression (`0xC0` pointers). Check against
   `dig`. Extra: if the `TC` bit is set, retry over TCP. TCP DNS prefixes each
   message with a 2-byte length, which is another framing rule.

6. **Exhaust ephemeral ports.** Run `gpcurl` against your local server 2,000
   times in a loop. While it runs, count connections in `TIME_WAIT`:
   `ss -tan state time-wait | wc -l`. Which side holds them, client or server?
   Then calculate: at 1,000 new connections per second with a 60 s
   `TIME_WAIT` and Linux's default port range, how long until a client runs
   out of ports to one destination?

7. **Feel the RTT.** Add 100 ms of delay to loopback:
   `sudo tc qdisc add dev lo root netem delay 50ms` (50 ms each way). Rerun
   `gpcurl` and see which phases grow, and by how much. Does it match the
   table in section 1? Remove it afterwards with
   `sudo tc qdisc del dev lo root`.

---

## 7. Project milestone M01: `gpcurl`

**Delivered in this module** (reference implementation in `gatepass/`):

- [x] Raw TCP client that sends a hand-written HTTP/1.1 request.
- [x] Its own response parser covering all four message-length rules,
      including chunked bodies and trailers.
- [x] Per-phase timings: DNS, TCP connect, TLS handshake, TTFB, content.
- [x] Works with `http://` and `https://`; verifies certificates; rejects
      untrusted ones.
- [x] Strict parsing: line and header limits, rejects ambiguous framing.
- [x] Tests against `net/http` servers, race detector clean.

**Acceptance check:**

```console
$ cd gatepass && go vet ./... && go test -race ./...
```

**Your extension tasks** (not in the reference code):

- Add `-I` (send `HEAD`, print headers only). Check the parser handles a
  `Content-Length` with no body.
- Add redirect following (`-L`) for `301/302/307/308`, with a limit of 10.
  Which status codes must keep the method and body?
- Try every resolved address in turn if connecting to the first one fails.

---

## 8. Checkpoint

You're ready to move on when you can answer these without notes:

1. List the round trips needed to fetch a page over a brand-new HTTPS
   connection with TLS 1.3. How many does a reused connection need?
2. What identifies a TCP connection? How can one server port serve a million
   clients?
3. Why does every protocol on top of TCP need its own framing rules?
4. Give the four HTTP/1.1 rules for finding the end of a response body, in
   order. What attack happens when two servers apply them differently?
5. What's the difference between flow control and congestion control? Which
   one stalls a sender when the receiving application reads slowly?
6. Why is a fresh TCP connection slower than a warm one even after the
   handshake?
7. What is `TIME_WAIT` for, which side ends up in it, and how can it break a
   busy client?
8. You change Gatepass's `A` record but some users still reach the old server
   an hour later. List three caches that could be responsible.
9. What problem does HTTP/2 solve that HTTP/1.1 keep-alive doesn't? What
   problem does HTTP/2 introduce that HTTP/3 solves?
10. What do SNI and ALPN tell the server, and why must they be sent before
    encryption starts?

---

## 9. Further reading

- Kurose & Ross, *Computer Networking: A Top-Down Approach*, chapters 2–3.
- Ilya Grigorik, [*High Performance Browser Networking*](https://hpbn.co/),
  chapters on TCP, TLS, HTTP/2 (free online).
- RFC 9110 (HTTP semantics) and RFC 9112 (HTTP/1.1). Read §6 of RFC 9112 on
  message bodies alongside `httpwire.go`.
- Julia Evans, *How DNS Works* (zine) and her blog posts on DNS.
- Cloudflare blog: "A Brief Primer on HTTP/3 and QUIC".
- `man 7 tcp` on Linux: every tunable discussed above.
