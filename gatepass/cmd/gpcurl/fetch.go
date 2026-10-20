package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nmutovkin/backend-course/gatepass/internal/platform/httpwire"
)

// Request is everything gpcurl needs to send one request.
type Request struct {
	Method string
	URL    *url.URL
	Header httpwire.Header // extra fields, sent after the defaults
	Body   string
}

// Timings records when each phase of a request finished. Phases that did
// not happen (DNS for an IP literal, TLS for http://) stay zero.
type Timings struct {
	Start, DNSDone, ConnectDone, TLSDone, FirstByte, Done time.Time
}

// Trace collects what happened on the wire, for printing in verbose mode.
type Trace struct {
	Addrs      []netip.Addr // every address DNS returned
	Remote     string       // the address we actually connected to
	Local      string       // our side of the connection (note the ephemeral port)
	TLS        *tls.ConnectionState
	RawRequest string
}

// Client fetches URLs over a raw TCP connection.
type Client struct {
	Resolver  *net.Resolver
	TLSConfig *tls.Config // nil means verify against the system roots
}

// Fetch sends req and writes the decoded response body to body.
// It always uses a new connection and asks the server to close it.
func (c *Client) Fetch(ctx context.Context, req Request, body io.Writer) (*httpwire.Response, *Timings, *Trace, error) {
	tm := &Timings{Start: time.Now()}
	tr := &Trace{}
	u := req.URL

	port := u.Port()
	switch {
	case port != "":
	case u.Scheme == "http":
		port = "80"
	case u.Scheme == "https":
		port = "443"
	default:
		return nil, tm, tr, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}

	// Phase 1: DNS. An IP literal needs no lookup.
	host := u.Hostname()
	if ip, err := netip.ParseAddr(host); err == nil {
		tr.Addrs = []netip.Addr{ip}
	} else {
		addrs, err := c.Resolver.LookupNetIP(ctx, "ip", host)
		if err != nil {
			return nil, tm, tr, fmt.Errorf("dns: %w", err)
		}
		tr.Addrs = addrs
		tm.DNSDone = time.Now()
	}

	// Phase 2: TCP connect (the three-way handshake happens inside Dial).
	// Real clients try every address ("Happy Eyeballs", RFC 8305); we take the first.
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(tr.Addrs[0].Unmap().String(), port))
	if err != nil {
		return nil, tm, tr, fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()
	tm.ConnectDone = time.Now()
	tr.Remote, tr.Local = conn.RemoteAddr().String(), conn.LocalAddr().String()
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}

	// Phase 3: TLS handshake, for https only.
	if u.Scheme == "https" {
		cfg := &tls.Config{}
		if c.TLSConfig != nil {
			cfg = c.TLSConfig.Clone()
		}
		cfg.ServerName = host                 // SNI, and the name the certificate must match
		cfg.NextProtos = []string{"http/1.1"} // ALPN: we only speak HTTP/1.1
		tc := tls.Client(conn, cfg)
		if err := tc.HandshakeContext(ctx); err != nil {
			return nil, tm, tr, fmt.Errorf("tls: %w", err)
		}
		state := tc.ConnectionState()
		tr.TLS = &state
		tm.TLSDone = time.Now()
		conn = tc
	}

	// Phase 4: send the request, then wait for the first byte of the answer.
	tr.RawRequest = buildRequest(req, host, port)
	if _, err := io.WriteString(conn, tr.RawRequest); err != nil {
		return nil, tm, tr, fmt.Errorf("write: %w", err)
	}
	fb := &firstByteReader{r: conn, at: &tm.FirstByte}
	resp, err := httpwire.ReadResponse(bufio.NewReader(fb), req.Method)
	if err != nil {
		return nil, tm, tr, fmt.Errorf("read: %w", err)
	}

	// Phase 5: content transfer.
	if _, err := io.Copy(body, resp.Body); err != nil {
		return resp, tm, tr, fmt.Errorf("read body: %w", err)
	}
	tm.Done = time.Now()
	return resp, tm, tr, nil
}

// buildRequest writes an HTTP/1.1 request by hand. Every line ends in CRLF
// and an empty line ends the header section.
func buildRequest(req Request, host, port string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s HTTP/1.1\r\n", req.Method, req.URL.RequestURI())

	// Host is the only mandatory header in HTTP/1.1. The port is omitted
	// when it is the scheme's default.
	hostHeader := req.URL.Hostname()
	if strings.Contains(hostHeader, ":") {
		hostHeader = "[" + hostHeader + "]" // IPv6 literal
	}
	if (req.URL.Scheme == "http" && port != "80") || (req.URL.Scheme == "https" && port != "443") {
		hostHeader += ":" + port
	}
	defaults := httpwire.Header{
		{Name: "Host", Value: hostHeader},
		{Name: "User-Agent", Value: "gpcurl/0.1"},
		{Name: "Accept", Value: "*/*"},
		{Name: "Connection", Value: "close"}, // one request per connection keeps reading simple
	}
	if req.Body != "" {
		defaults = append(defaults, httpwire.Field{Name: "Content-Length", Value: strconv.Itoa(len(req.Body))})
	}
	for _, f := range append(defaults, req.Header...) {
		fmt.Fprintf(&b, "%s: %s\r\n", f.Name, f.Value)
	}
	b.WriteString("\r\n")
	b.WriteString(req.Body)
	return b.String()
}

// firstByteReader records the moment the first byte arrives.
type firstByteReader struct {
	r  io.Reader
	at *time.Time
}

func (f *firstByteReader) Read(p []byte) (int, error) {
	n, err := f.r.Read(p)
	if n > 0 && f.at.IsZero() {
		*f.at = time.Now()
	}
	return n, err
}
