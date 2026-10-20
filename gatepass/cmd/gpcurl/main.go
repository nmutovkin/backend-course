// Command gpcurl is a tiny HTTP/1.1 client built on a raw TCP socket.
// It prints how long each phase of a request took: DNS, TCP connect,
// TLS handshake, waiting for the server, and content transfer.
//
//	gpcurl [-v] [-X method] [-H 'Name: value']... [-d body] URL
//
// The response body goes to stdout; status, headers and timings go to
// stderr, so `gpcurl URL > page.html` still shows the timings.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/nmutovkin/backend-course/gatepass/internal/platform/httpwire"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

type headerFlags httpwire.Header

func (h *headerFlags) String() string { return "" }
func (h *headerFlags) Set(v string) error {
	name, value, ok := strings.Cut(v, ":")
	if !ok {
		return fmt.Errorf("header %q must look like 'Name: value'", v)
	}
	*h = append(*h, httpwire.Field{Name: strings.TrimSpace(name), Value: strings.TrimSpace(value)})
	return nil
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gpcurl", flag.ContinueOnError)
	fs.SetOutput(stderr)
	method := fs.String("X", "GET", "request method")
	data := fs.String("d", "", "request body")
	verbose := fs.Bool("v", false, "print addresses, TLS details and the raw request")
	timeout := fs.Duration("timeout", 30*time.Second, "overall deadline")
	var headers headerFlags
	fs.Var(&headers, "H", "extra header 'Name: value' (repeatable)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: gpcurl [flags] URL")
		return 2
	}
	u, err := url.Parse(fs.Arg(0))
	if err != nil || u.Host == "" {
		fmt.Fprintf(stderr, "gpcurl: invalid URL %q\n", fs.Arg(0))
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	c := &Client{Resolver: net.DefaultResolver}
	req := Request{Method: strings.ToUpper(*method), URL: u, Header: httpwire.Header(headers), Body: *data}

	resp, tm, tr, err := c.Fetch(ctx, req, stdout)
	if *verbose {
		printTrace(stderr, tr)
	}
	if resp != nil {
		fmt.Fprintf(stderr, "< %s %d %s\n", resp.Proto, resp.StatusCode, resp.Reason)
		for _, f := range resp.Header {
			fmt.Fprintf(stderr, "< %s: %s\n", f.Name, f.Value)
		}
	}
	printTimings(stderr, tm)
	if err != nil {
		fmt.Fprintf(stderr, "gpcurl: %v\n", err)
		return 1
	}
	return 0
}

func printTrace(w io.Writer, tr *Trace) {
	for _, a := range tr.Addrs {
		fmt.Fprintf(w, "* resolved %s\n", a.Unmap())
	}
	if tr.Remote != "" {
		fmt.Fprintf(w, "* connected %s -> %s\n", tr.Local, tr.Remote)
	}
	if tr.TLS != nil {
		fmt.Fprintf(w, "* %s, %s, ALPN %q\n", tls.VersionName(tr.TLS.Version),
			tls.CipherSuiteName(tr.TLS.CipherSuite), tr.TLS.NegotiatedProtocol)
		if certs := tr.TLS.PeerCertificates; len(certs) > 0 {
			leaf := certs[0]
			fmt.Fprintf(w, "* certificate %s, issued by %s, expires %s\n",
				leaf.Subject.CommonName, leaf.Issuer.CommonName, leaf.NotAfter.Format(time.DateOnly))
		}
	}
	if tr.RawRequest == "" {
		return
	}
	for _, line := range strings.Split(strings.TrimRight(tr.RawRequest, "\r\n"), "\r\n") {
		fmt.Fprintf(w, "> %s\n", line)
	}
}

// printTimings prints each phase as the time since the previous phase
// finished. Phases that did not happen are skipped.
func printTimings(w io.Writer, tm *Timings) {
	prev := tm.Start
	phase := func(name string, at time.Time) {
		if at.IsZero() {
			return
		}
		fmt.Fprintf(w, "  %-16s %8s\n", name, at.Sub(prev).Round(time.Microsecond))
		prev = at
	}
	fmt.Fprintln(w, "timings:")
	phase("DNS lookup", tm.DNSDone)
	phase("TCP connect", tm.ConnectDone)
	phase("TLS handshake", tm.TLSDone)
	phase("Server (TTFB)", tm.FirstByte)
	phase("Content", tm.Done)
	if !tm.Done.IsZero() {
		fmt.Fprintf(w, "  %-16s %8s\n", "Total", tm.Done.Sub(tm.Start).Round(time.Microsecond))
	}
}
