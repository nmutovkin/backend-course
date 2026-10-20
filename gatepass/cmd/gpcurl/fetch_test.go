package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// chunkedHandler flushes between writes, which makes net/http send the
// body with Transfer-Encoding: chunked.
func chunkedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Seen-Host", r.Host)
	w.Header().Set("X-Seen-Method", r.Method)
	body, _ := io.ReadAll(r.Body)
	io.WriteString(w, "first,")
	w.(http.Flusher).Flush()
	io.WriteString(w, "second,")
	w.(http.Flusher).Flush()
	io.WriteString(w, "echo="+string(body))
}

func fetch(t *testing.T, c *Client, method, rawURL, body string) (string, http.Header) {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	resp, tm, _, err := c.Fetch(context.Background(), Request{Method: method, URL: u, Body: body}, &out)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if tm.ConnectDone.IsZero() || tm.FirstByte.IsZero() || tm.Done.IsZero() {
		t.Errorf("missing timings: %+v", tm)
	}
	h := http.Header{}
	for _, f := range resp.Header {
		h.Add(f.Name, f.Value)
	}
	return out.String(), h
}

func TestFetchHTTPChunked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(chunkedHandler))
	defer srv.Close()

	body, h := fetch(t, &Client{Resolver: net.DefaultResolver}, "POST", srv.URL+"/events?city=lisbon", "hi")
	if want := "first,second,echo=hi"; body != want {
		t.Errorf("body = %q, want %q", body, want)
	}
	if h.Get("Transfer-Encoding") != "chunked" {
		t.Errorf("expected a chunked response, got headers %v", h)
	}
	if h.Get("X-Seen-Method") != "POST" {
		t.Errorf("server saw method %q", h.Get("X-Seen-Method"))
	}
	if want := strings.TrimPrefix(srv.URL, "http://"); h.Get("X-Seen-Host") != want {
		t.Errorf("server saw Host %q, want %q", h.Get("X-Seen-Host"), want)
	}
}

func TestFetchHTTPS(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(chunkedHandler))
	defer srv.Close()

	// Trust the test server's self-signed certificate. It is issued for
	// "example.com" and 127.0.0.1, so connecting by IP verifies cleanly.
	roots := x509.NewCertPool()
	roots.AddCert(srv.Certificate())
	c := &Client{Resolver: net.DefaultResolver, TLSConfig: &tls.Config{RootCAs: roots}}

	body, _ := fetch(t, c, "GET", srv.URL, "")
	if want := "first,second,echo="; body != want {
		t.Errorf("body = %q, want %q", body, want)
	}
}

func TestFetchRejectsUntrustedCertificate(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(chunkedHandler))
	defer srv.Close()

	u, _ := url.Parse(srv.URL)
	_, _, _, err := (&Client{Resolver: net.DefaultResolver}).Fetch(context.Background(), Request{Method: "GET", URL: u}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "tls:") {
		t.Fatalf("err = %v, want a TLS verification error", err)
	}
}

func TestBuildRequest(t *testing.T) {
	u, _ := url.Parse("http://gatepass.test:8080/events/42?seats=A1,A2")
	got := buildRequest(Request{Method: "GET", URL: u}, "gatepass.test", "8080")
	want := "GET /events/42?seats=A1,A2 HTTP/1.1\r\n" +
		"Host: gatepass.test:8080\r\n" +
		"User-Agent: gpcurl/0.1\r\n" +
		"Accept: */*\r\n" +
		"Connection: close\r\n" +
		"\r\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}
