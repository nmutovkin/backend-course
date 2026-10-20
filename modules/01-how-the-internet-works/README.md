# Module 01: How the Internet Works

> **Depth:** Transport (TCP/UDP), DNS and HTTP at `L3`; networking models and
> addressing at `L2`; TLS, browsers and hosting at `L1` (TLS returns at `L2`
> in Module 12).
>
> **You will build:** `gpcurl`, an HTTP client on a raw TCP socket that
> shows where every millisecond of a request goes.
>
> **Time:** ~12 hours.
>
> **Before you start:** finish [Module 00](../00-setup-and-go-primer/README.md).
> You need Go, the network tools from its §2.3, and enough Go to read the code.

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

**On the wire.** A DNS message is a compact binary format (RFC 1035). It
normally travels as a single UDP datagram to port 53. Everything below is
enough to build a working DNS client (Lab 5).

All multi-byte integers are **big-endian** ("network byte order"): the most
significant byte comes first, so the 16-bit number `0x0100` is the bytes
`01 00`.

A message has a 12-byte header, then four sections: question, answer,
authority and additional.

**Header (12 bytes):**

| Bytes | Field     | Meaning |
|-------|-----------|---------|
| 0–1   | `ID`      | Random number chosen by the client; the reply carries the same ID |
| 2–3   | flags     | 16 bits, from the most significant bit: `QR`(1) `Opcode`(4) `AA`(1) `TC`(1) `RD`(1) `RA`(1) `Z`(3) `RCODE`(4) |
| 4–5   | `QDCOUNT` | number of questions (always 1 in practice) |
| 6–7   | `ANCOUNT` | number of answer records |
| 8–9   | `NSCOUNT` | number of authority records |
| 10–11 | `ARCOUNT` | number of additional records |

Flag meanings: `QR` 0 = query, 1 = response. `Opcode` 0 = standard query.
`AA` authoritative answer. `TC` truncated. `RD` recursion desired (the client
asks the resolver to do the whole walk). `RA` recursion available. `Z` must
be zero. `RCODE` result: 0 `NOERROR`, 2 `SERVFAIL`, 3 `NXDOMAIN` (name
doesn't exist), 5 `REFUSED`.

So a query with only `RD` set has flags `0x0100`. A typical successful
recursive response has `0x8180` (`QR`, `RD`, `RA`).

**Names** are encoded as a sequence of *labels*. Each label is one length
byte (1–63) followed by that many characters, and a zero byte ends the name:

```
api.gatepass.test  →  03 'a' 'p' 'i' 08 'g' 'a' 't' 'e' 'p' 'a' 's' 's' 04 't' 'e' 's' 't' 00
```

**Question:** `QNAME` (an encoded name), then `QTYPE` (2 bytes) and `QCLASS`
(2 bytes, always 1 = `IN`, Internet). Type codes: `A`=1, `NS`=2, `CNAME`=5,
`SOA`=6, `MX`=15, `TXT`=16, `AAAA`=28, `SRV`=33, `CAA`=257.

**Resource record** (each entry in the answer, authority and additional
sections):

| Field      | Size     | Meaning |
|------------|----------|---------|
| `NAME`     | variable | encoded name (often compressed, see below) |
| `TYPE`     | 2 bytes  | type code, as above |
| `CLASS`    | 2 bytes  | 1 = `IN` |
| `TTL`      | 4 bytes  | seconds this record may be cached (unsigned) |
| `RDLENGTH` | 2 bytes  | length of `RDATA` in bytes |
| `RDATA`    | `RDLENGTH` bytes | the data: 4 bytes for `A`, 16 for `AAAA`, an encoded name for `CNAME`/`NS` |

**Name compression.** Names repeat a lot in a response, so a name can end
with a **pointer** to a name earlier in the message. A length byte whose top
two bits are both 1 (any value `0xC0` or above) isn't a length. Together
with the next byte it forms a 14-bit offset from the start of the message:
`offset = (b1 & 0x3F) << 8 | b2`. The reader continues reading labels at that
offset, and the name ends at the pointer in its original position. The
question name always starts at offset 12, right after the header, so answers
usually begin with `C0 0C`: "same name as the question". Pointers can chain,
so a careful reader limits how many it follows, to survive malicious loops.

**A complete example.** Query for the `A` record of `example.com`, ID
`0xBEEF`, recursion desired (29 bytes):

```
be ef  01 00  00 01  00 00  00 00  00 00       header: ID, flags=RD, QDCOUNT=1
07 65 78 61 6d 70 6c 65  03 63 6f 6d  00       "example" "com" end
00 01  00 01                                   QTYPE=A, QCLASS=IN
```

A response to it, with one answer `203.0.113.7` and a TTL of 3600:

```
be ef  81 80  00 01  00 01  00 00  00 00       same ID, flags=QR|RD|RA, 1 question, 1 answer
07 65 78 61 6d 70 6c 65  03 63 6f 6d  00  00 01 00 01    question echoed back
c0 0c                                          NAME: pointer to offset 12 (example.com)
00 01  00 01                                   TYPE=A, CLASS=IN
00 00 0e 10                                    TTL = 0x0e10 = 3600 s
00 04                                          RDLENGTH = 4
cb 00 71 07                                    203.0.113.7
```

Practical details:

- Responses over plain UDP are limited to 512 bytes. **EDNS(0)** lets the
  client advertise a larger buffer through an extra record in the additional
  section.
- If the answer still doesn't fit, the server sets **TC** and the client
  repeats the query over **TCP**, port 53. Over TCP, each message is prefixed
  with its length as 2 bytes, because TCP has no message boundaries
  (section 2.3): framing again.
- Randomizing the ID and the UDP source port is what makes forging responses
  ("cache poisoning") hard: an attacker would have to guess both.
- **DoT** and **DoH** (DNS over TLS and over HTTPS) send these same messages
  inside an encrypted connection.

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

### Lab setup

Most labs need a local web server and a built `gpcurl`. Open a terminal and
start the server. It keeps running, so leave this terminal alone:

```console
$ mkdir -p /tmp/www && cd /tmp/www
$ echo "hello from riverside" > index.html
$ python3 -m http.server 8099
Serving HTTP on 0.0.0.0 port 8099 (http://0.0.0.0:8099/) ...
```

`python3 -m http.server 8099` serves the current directory over HTTP on port
8099. It speaks HTTP/1.0 and closes the connection after every response. In
a second terminal, build `gpcurl` once:

```console
$ cd backend-course/gatepass
$ go build -o /tmp/gpcurl ./cmd/gpcurl
$ /tmp/gpcurl http://127.0.0.1:8099/
hello from riverside
```

Stop the server with `Ctrl+C` when you're done with the labs.

---

### Lab 1: Watch the handshake

**Goal:** see section 3.1 happen, packet by packet.

1. In a third terminal, start a packet capture on the loopback interface:
   ```console
   $ sudo tcpdump -i lo -n -S 'tcp port 8099'
   ```
   - `-i lo`: capture on the loopback interface (macOS: `-i lo0`).
   - `-n`: print addresses as numbers instead of looking up names.
   - `-S`: print absolute sequence numbers, so you can do the arithmetic.
   - `'tcp port 8099'`: a filter, so you only see our traffic.
2. In the second terminal run `/tmp/gpcurl http://127.0.0.1:8099/`.
3. Stop tcpdump with `Ctrl+C`.

**What you should see** (TCP options trimmed; your ports, sequence numbers and
times will differ):

```
IP 127.0.0.1.33578 > 127.0.0.1.8099: Flags [S],  seq 908084557, length 0
IP 127.0.0.1.8099 > 127.0.0.1.33578: Flags [S.], seq 2127422235, ack 908084558, length 0
IP 127.0.0.1.33578 > 127.0.0.1.8099: Flags [.],  ack 2127422236, length 0
IP 127.0.0.1.33578 > 127.0.0.1.8099: Flags [P.], seq 908084558:908084654, ack 2127422236, length 96
IP 127.0.0.1.8099 > 127.0.0.1.33578: Flags [.],  ack 908084654, length 0
IP 127.0.0.1.8099 > 127.0.0.1.33578: Flags [P.], seq 2127422236:2127422422, ack 908084654, length 186
IP 127.0.0.1.33578 > 127.0.0.1.8099: Flags [.],  ack 2127422422, length 0
IP 127.0.0.1.8099 > 127.0.0.1.33578: Flags [P.], seq 2127422422:2127422443, ack 908084654, length 21
IP 127.0.0.1.33578 > 127.0.0.1.8099: Flags [.],  ack 2127422443, length 0
IP 127.0.0.1.8099 > 127.0.0.1.33578: Flags [F.], seq 2127422443, ack 908084654, length 0
IP 127.0.0.1.33578 > 127.0.0.1.8099: Flags [F.], seq 908084654, ack 2127422444, length 0
IP 127.0.0.1.8099 > 127.0.0.1.33578: Flags [.],  ack 908084655, length 0
```

How to read it. `Flags` shows the TCP flags: `S` = SYN, `F` = FIN,
`R` = RST, `P` = PSH ("deliver this to the application now"), and `.` = ACK.
So `[S.]` is SYN+ACK. `seq a:b` covers bytes `a` up to (not including) `b`.

- Lines 1–3 are the **three-way handshake**. The server's `ack 908084558` is
  the client's `seq + 1`: a SYN counts as one byte.
- Line 4 is our request: **96 bytes**. Count it from `buildRequest`'s output:
  16 + 22 + 24 + 13 + 19 + 2 bytes for the six lines, CRLFs included.
- Lines 6 and 8 are the response headers (186 bytes) and the body (21 bytes,
  `hello from riverside\n`). Python writes them separately, so they travel in
  two segments.
- Lines 10–12 close the connection, one direction at a time.

**Questions:**

<details>
<summary>Which side sent the first FIN, and why?</summary>

The server (line 10). It speaks HTTP/1.0 and we sent `Connection: close`, so
it closes once the response is written. The side that closes first goes into
`TIME_WAIT` (section 3.1), so here that's the server. Lab 6 confirms it.
</details>

<details>
<summary>Why are there separate ACK-only segments (length 0)?</summary>

Each side acknowledges data it received. When the receiver has nothing to
send right away, the ACK goes out on its own. Delayed ACK (section 3.1) tries
to combine ACKs with data, but here both sides are fast enough that
standalone ACKs appear.
</details>

<details>
<summary>The SYN's options include <code>mss 65495</code>. Why so large?</summary>

MSS follows the interface's maximum packet size (MTU). Loopback's MTU is
65,536 bytes; Ethernet's is 1,500, which gives an MSS of 1,460. On a real
network you would see 1460 or less.
</details>

---

### Lab 2: Speak HTTP by hand

**Goal:** see that HTTP/1.1 is just text, and how a server reacts to bad input.

`nc` (netcat) opens a TCP connection and copies your keyboard to it and the
reply to your screen. `printf` lets us send exact bytes, including `\r\n`:

```console
$ printf 'GET / HTTP/1.1\r\nHost: localhost\r\n\r\n' | nc 127.0.0.1 8099
HTTP/1.0 200 OK
Server: SimpleHTTP/0.6 Python/3.13.16
Date: Tue, 20 Oct 2026 17:17:10 GMT
Content-type: text/html
Content-Length: 21
Last-Modified: Tue, 20 Oct 2026 17:17:06 GMT

hello from riverside
```

Now send each of these and look at the first line of the reply:

| Send                                              | Python's reply                          |
|---------------------------------------------------|-----------------------------------------|
| `'GET / HTTP/1.1\r\n\r\n'` (no `Host`)            | `HTTP/1.0 200 OK`                       |
| `'GET / HTTP/1.0\r\n\r\n'`                        | `HTTP/1.0 200 OK`                       |
| `'FOO / HTTP/1.1\r\nHost: x\r\n\r\n'`             | `HTTP/1.0 501 Unsupported method ('FOO')` |
| `'GET /nope HTTP/1.1\r\nHost: x\r\n\r\n'`         | `HTTP/1.0 404 File not found`           |
| `'hello\r\n\r\n'`                                 | `HTTP/1.0 400 Bad request syntax ('hello')` |

<details>
<summary>The first row is a bug in Python's server. Why?</summary>

HTTP/1.1 requires a `Host` header, and RFC 9112 says a server *must* answer
`400 Bad Request` when it is missing. Python's development server is lenient.
Go's `net/http` and nginx both reject it with 400. Leniency like this is
exactly what request smuggling exploits (section 3.3), which is why the
server we build in Module 05 will be strict.
</details>

<details>
<summary>Why does the server answer <code>HTTP/1.0</code> when we asked in 1.1?</summary>

A server replies with the highest version it supports, up to the client's.
Python's simple server only implements 1.0, so it answers in 1.0, which also
means it closes the connection after each response.
</details>

---

### Lab 3: Follow the DNS chain

**Goal:** watch a recursive lookup (section 3.2). Needs internet access.

```console
$ dig +trace example.com
```

`dig` sends DNS queries and prints the answers in a readable form. `+trace`
makes `dig` do the recursion itself, starting at the root, instead of asking
your resolver. The output comes in four blocks, roughly:

```
.                       518400  IN  NS  a.root-servers.net.     ← from your resolver: who are the roots
...
com.                    172800  IN  NS  a.gtld-servers.net.     ← a root server: "ask the .com servers"
...
example.com.            172800  IN  NS  a.iana-servers.net.     ← a .com server: "ask example.com's servers"
...
example.com.            3600    IN  A   <address>               ← the authoritative answer
```

Each line is one record: name, TTL in seconds, class, type, data. The
comment line under each block (`;; Received 525 bytes from 198.41.0.4#53(a.root-servers.net)`)
says which server answered.

Now ask your normal resolver twice, a few seconds apart:

```console
$ dig example.com A +noall +answer
example.com.    3600    IN  A   <address>
$ dig example.com A +noall +answer
example.com.    3593    IN  A   <address>
```

(`+noall +answer` prints only the answer section.)

<details>
<summary>Why did the TTL go down between the two queries?</summary>

The second answer came from the resolver's cache. The resolver counts the
TTL down while it caches the record, so you see how long it has left. When
it reaches 0, the resolver asks the authoritative server again. If the first
TTL was already below the record's real TTL, someone else had recently asked
your resolver the same question.
</details>

---

### Lab 4: Compare with curl

**Goal:** check `gpcurl`'s timings against an independent tool.

curl can print timings with `-w` (write-out):

```console
$ curl -o /dev/null -s -w 'dns %{time_namelookup} connect %{time_connect} tls %{time_appconnect} ttfb %{time_starttransfer} total %{time_total}\n' https://example.com
$ /tmp/gpcurl https://example.com > /dev/null
```

- `-o /dev/null`: throw the body away. `-s`: silent, no progress bar.
- curl's numbers are **cumulative seconds since the start**: `connect` is
  "DNS + TCP". gpcurl's are **per phase**.

To compare: gpcurl's *TCP connect* ≈ curl's `connect − dns`; *TLS handshake*
≈ `tls − connect`; *Server (TTFB)* ≈ `ttfb − tls`.

Then run `ping -c 5 example.com` and compare its average RTT with the
connect phase.

<details>
<summary>Which phase should be closest to one RTT, and why?</summary>

TCP connect: it's exactly one round trip (SYN out, SYN-ACK back). The TLS 1.3
handshake is also about one RTT plus some cryptography time. TTFB is one RTT
plus however long the server takes. Without internet access, run both tools
against `http://127.0.0.1:8099/`. The RTT is then a few microseconds, and
the phases mostly show local processing time.
</details>

---

### Lab 5 (stretch, `L3` DNS): build `gpdig`

**Goal:** a DNS client written from the wire format, with no DNS library.
Section 3.2 contains everything you need.

Work in a scratch module, outside `gatepass/`:

```console
$ mkdir -p ~/goplay/gpdig && cd ~/goplay/gpdig && go mod init gpdig
```

Suggested steps:

1. **`buildQuery(id uint16, name string, qtype uint16) []byte`.** Write the
   header with `binary.BigEndian.PutUint16` (package `encoding/binary`), then
   the encoded name, then `QTYPE` and `QCLASS`. **Test it:** for
   `0xBEEF, "example.com", 1` it must return exactly the 29 query bytes from
   section 3.2.
2. **`readName(msg []byte, off int) (name string, next int, err error)`.**
   Read labels from `off`. On a byte `>= 0xC0`, follow the pointer, and
   remember that `next` is the position right after the 2-byte pointer, not
   wherever the pointer led. Give up after, say, 10 pointers.
   Return an error rather than reading past the end of `msg`.
3. **`parseResponse(msg []byte) ([]Answer, error)`.** Check the ID matches and
   `QR` is set, read `RCODE` (`flags & 0x000F`), skip the question (a name,
   then 4 bytes), then read `ANCOUNT` records. For type 1, `RDATA` is the
   IPv4 address (`net.IP(rdata).String()`). **Test it** with the
   response bytes from section 3.2: you should get one answer,
   `203.0.113.7`, TTL 3600.
4. **Send it.** Find your resolver with `grep nameserver /etc/resolv.conf`,
   then:
   ```go
   conn, err := net.Dial("udp", "RESOLVER_IP:53") // a UDP "connection" just fixes the peer address
   conn.SetDeadline(time.Now().Add(3 * time.Second))
   conn.Write(query)
   buf := make([]byte, 512)
   n, err := conn.Read(buf) // one datagram = one whole DNS message
   ```
   Pick the ID with `uint16(rand.Uint32())` (package `math/rand/v2`).
5. Compare your output with `dig example.com A +noall +answer`.

**Extensions:** support `AAAA` (type 28, 16-byte address) and `CNAME`
(type 5, `RDATA` is a possibly-compressed name, so use `readName`). If `TC`
(`flags & 0x0200`) is set, retry over TCP with the 2-byte length prefix.
Print `NXDOMAIN` when `RCODE` is 3.

---

### Lab 6: Exhaust ephemeral ports (Linux)

**Goal:** see `TIME_WAIT` pile up (section 3.1).

With the lab server running:

```console
$ for i in $(seq 2000); do /tmp/gpcurl http://127.0.0.1:8099/ >/dev/null 2>&1; done
$ ss -tan state time-wait | head -3
Recv-Q Send-Q Local Address:Port   Peer Address:Port
0      0          127.0.0.1:8099      127.0.0.1:34684
0      0          127.0.0.1:8099      127.0.0.1:34000
$ ss -tan state time-wait | wc -l
2001              ← about one per request, plus the header line
```

`ss` lists sockets: `-t` TCP only, `-a` all states, `-n` numeric. `state
time-wait` filters by state. The count includes one header line. (macOS:
`netstat -an -p tcp | grep TIME_WAIT | wc -l`.) Wait 60 seconds and count
again: they're gone.

<details>
<summary>Which side holds the TIME_WAIT sockets here? Would it be the same against a real HTTP/1.1 server?</summary>

The server: the local address is `127.0.0.1:8099`, because Python closes
first (Lab 1). A keep-alive HTTP/1.1 server usually waits for the client to
close, and then the *client* collects `TIME_WAIT` sockets. That's the
dangerous case, because the client is the one that needs ephemeral ports.
</details>

<details>
<summary>At 1,000 new connections per second to one server, how long until the client runs out of ports?</summary>

Linux's default range 32768–60999 holds 28,232 ports. Each closed connection
keeps its port in `TIME_WAIT` for 60 s, so at most 28,232 / 60 ≈ 470 new
connections per second can be sustained to one destination IP and port. At
1,000 per second the ports run out after about 28 seconds, and connects
start failing with `EADDRNOTAVAIL`. Fixes, from best to worst: reuse
connections (a connection pool), spread over more destination addresses, and
widen the port range. Linux can also reuse `TIME_WAIT` ports for outgoing
connections when TCP timestamps are on (`net.ipv4.tcp_tw_reuse`), but that
only treats the symptom.
</details>

---

### Lab 7: Feel the RTT (Linux)

**Goal:** check the round-trip table from section 1 by adding latency.

`tc` (traffic control) can delay every packet on an interface with the
`netem` (network emulator) module. Every packet on loopback crosses it once,
so a 50 ms delay makes the RTT 100 ms:

```console
$ sudo tc qdisc add dev lo root netem delay 50ms
$ /tmp/gpcurl http://127.0.0.1:8099/ > /dev/null
$ sudo tc qdisc del dev lo root          # always remove it afterwards
```

(If you get `Specified qdisc kind is unknown`, your kernel lacks the
`netem` module. Some container and WSL kernels do. Try
`sudo modprobe sch_netem`, or use a full Linux VM.)

<details>
<summary>What should the timings show?</summary>

No DNS phase (an IP literal). *TCP connect* ≈ 100 ms: one RTT. *Server
(TTFB)* ≈ 100 ms: one RTT for the request to arrive and the response to come
back. *Content* close to 0, because the body follows the headers right away
and both are delayed by the same 50 ms. Total ≈ 200 ms: two round trips
for plain HTTP. With https, add another ≈100 ms for the TLS handshake,
matching section 1's table.
</details>

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
- Add redirect following (`-L`) for `301/302/303/307/308`, with a limit of
  10. The new URL is in the `Location` header and may be relative: resolve it
  against the current URL with `u.ResolveReference(loc)`. Method rules:
  `307` and `308` must repeat the same method and body. `303` always switches
  to `GET` without a body. `301` and `302` should keep the method, but for
  historical reasons every real client turns a `POST` into a `GET`, so do
  the same. (Module 05 covers status codes in full.)
- Try every resolved address in turn if connecting to the first one fails.

---

## 8. Checkpoint

You're ready to move on when you can answer these without notes. Answer
first, then open the answer to check.

<details>
<summary>1. List the round trips needed to fetch a page over a brand-new HTTPS connection with TLS 1.3. How many does a reused connection need?</summary>

DNS (one to the resolver, if not cached), one for the TCP handshake, one for
TLS 1.3, one for the request and response: about 4 round trips, about 3 if
DNS is cached. A reused connection needs only the request round trip:
one. (Section 1.)
</details>

<details>
<summary>2. What identifies a TCP connection? How can one server port serve a million clients?</summary>

The 5-tuple: protocol, source IP, source port, destination IP, destination
port. All connections to `:443` share the destination half but differ in the
client's IP or ephemeral port, so each is distinct. The real limits are
memory and file descriptors (Module 02), not ports. (Section 2.3.)
</details>

<details>
<summary>3. Why does every protocol on top of TCP need its own framing rules?</summary>

TCP delivers a byte stream, not messages. One write may arrive in pieces or
merged with the next, so the receiver needs rules (a length prefix, a
delimiter, chunking) to find where each message ends. (Section 2.3.)
</details>

<details>
<summary>4. Give the four HTTP/1.1 rules for finding the end of a response body, in order. What attack happens when two servers apply them differently?</summary>

(1) `HEAD` responses and `1xx`/`204`/`304` have no body. (2)
`Transfer-Encoding: chunked`: read chunks until the zero-size one; this wins
over `Content-Length`. (3) `Content-Length: N`: read exactly N bytes. (4)
Otherwise, read until the connection closes. If a proxy and a backend
disagree, an attacker can hide a second request inside the first one's body:
HTTP request smuggling. (Section 3.3.)
</details>

<details>
<summary>5. What's the difference between flow control and congestion control? Which one stalls a sender when the receiving application reads slowly?</summary>

Flow control protects the receiver: the receive window `rwnd` says how much
buffer space is left. Congestion control protects the network: the
congestion window `cwnd` grows and shrinks with observed loss. A slowly
reading application fills its receive buffer, `rwnd` drops to zero, and
flow control stalls the sender. (Section 3.1.)
</details>

<details>
<summary>6. Why is a fresh TCP connection slower than a warm one even after the handshake?</summary>

Slow start: a new connection begins with a small congestion window (about
14 KB) and doubles it every RTT, so a large response needs several round
trips before it reaches full speed. A warm connection already has a large
`cwnd`. (Section 3.1.)
</details>

<details>
<summary>7. What is <code>TIME_WAIT</code> for, which side ends up in it, and how can it break a busy client?</summary>

It keeps a closed connection's 5-tuple reserved for about 60 s, so that
delayed packets from the old connection can't be mistaken for a new one.
The side that closes first enters it. A client that opens and closes many
connections to one server keeps every port in `TIME_WAIT`, runs out of
ephemeral ports, and new connects fail with `EADDRNOTAVAIL`. The fix is
connection reuse. (Section 3.1, Lab 6.)
</details>

<details>
<summary>8. You change Gatepass's <code>A</code> record but some users still reach the old server an hour later. List three caches that could be responsible.</summary>

The recursive resolver (ISP, `8.8.8.8`) holding the record until its TTL
expires; the operating system's cache (`systemd-resolved`, `nscd`); the
browser's or application runtime's own DNS cache. Also long-lived
connections, which don't look DNS up again at all. Lowering the TTL a day
before the change is the prevention. (Section 2.4, §3.2.)
</details>

<details>
<summary>9. What problem does HTTP/2 solve that HTTP/1.1 keep-alive doesn't? What problem does HTTP/2 introduce that HTTP/3 solves?</summary>

HTTP/1.1 allows one request in flight per connection, so a slow response
blocks the rest (head-of-line blocking at the HTTP level). HTTP/2 runs many
streams over one connection. But on one TCP connection a single lost packet
stalls every stream until it's retransmitted (head-of-line blocking at the
TCP level). HTTP/3 runs over QUIC, whose streams are independent, so a loss
only stalls its own stream. (Sections 3.3, 3.4.)
</details>

<details>
<summary>10. What do SNI and ALPN tell the server, and why must they be sent before encryption starts?</summary>

SNI names the hostname the client wants. ALPN lists the application
protocols it speaks (`h2`, `http/1.1`). The server needs both to pick the
right certificate and protocol, and it must do that during the handshake,
before any keys exist. That's why they're sent in the clear, in the first
handshake message. (Section 2.6.)
</details>

---

## 9. Further reading (optional)

Everything you need is above. These go deeper:

- Kurose & Ross, *Computer Networking: A Top-Down Approach*, chapters 2–3.
- Ilya Grigorik, [*High Performance Browser Networking*](https://hpbn.co/),
  chapters on TCP, TLS, HTTP/2 (free online).
- RFC 9110 (HTTP semantics) and RFC 9112 (HTTP/1.1). Read §6 of RFC 9112 on
  message bodies alongside `httpwire.go`.
- Julia Evans, *How DNS Works* (zine) and her blog posts on DNS.
- Cloudflare blog: "A Brief Primer on HTTP/3 and QUIC".
- `man 7 tcp` on Linux: every tunable discussed above.
