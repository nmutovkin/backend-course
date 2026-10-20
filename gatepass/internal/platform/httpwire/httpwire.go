// Package httpwire implements just enough of the HTTP/1.1 wire format
// (RFC 9112) to be read in one sitting. Module 01 uses it to read
// responses; Module 05 adds the server side (requests).
//
// It is deliberately small and strict. Production code should use the
// standard library's net/http, which handles far more edge cases.
package httpwire

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Limits protect the reader from peers that send endless lines or headers.
const (
	MaxLineLength  = 8 << 10 // 8 KiB per status line, header line or chunk-size line
	MaxHeaderCount = 100
)

var (
	ErrLineTooLong    = errors.New("httpwire: line too long")
	ErrTooManyHeaders = errors.New("httpwire: too many header fields")
	ErrMalformed      = errors.New("httpwire: malformed message")
)

// Field is one header field. A Header keeps fields in the order they
// arrived and keeps duplicates, exactly as they were on the wire.
type Field struct {
	Name  string
	Value string
}

// Header is an ordered list of header fields.
type Header []Field

// Get returns the value of the first field named name (case-insensitive),
// or "" if there is none.
func (h Header) Get(name string) string {
	for _, f := range h {
		if strings.EqualFold(f.Name, name) {
			return f.Value
		}
	}
	return ""
}

// Values returns the values of every field named name (case-insensitive).
func (h Header) Values(name string) []string {
	var vs []string
	for _, f := range h {
		if strings.EqualFold(f.Name, name) {
			vs = append(vs, f.Value)
		}
	}
	return vs
}

// Response is a parsed HTTP/1.x response. Body yields the decoded body
// bytes (chunked framing removed) and returns io.EOF at its end.
type Response struct {
	Proto      string // e.g. "HTTP/1.1"
	StatusCode int    // e.g. 200
	Reason     string // e.g. "OK"; may be empty
	Header     Header
	Body       io.Reader
}

// ReadResponse reads one response from r. method is the method of the
// request it answers: a response to HEAD never has a body, whatever its
// headers say.
func ReadResponse(r *bufio.Reader, method string) (*Response, error) {
	line, err := readLine(r)
	if err != nil {
		return nil, fmt.Errorf("reading status line: %w", err)
	}
	resp, err := parseStatusLine(line)
	if err != nil {
		return nil, err
	}
	if resp.Header, err = readHeader(r); err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}
	if resp.Body, err = bodyReader(r, method, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// parseStatusLine parses "HTTP/1.1 200 OK".
func parseStatusLine(line string) (*Response, error) {
	proto, rest, ok := strings.Cut(line, " ")
	if !ok || !strings.HasPrefix(proto, "HTTP/1.") {
		return nil, fmt.Errorf("%w: status line %q", ErrMalformed, line)
	}
	code, reason, _ := strings.Cut(rest, " ")
	n, err := strconv.Atoi(code)
	if err != nil || len(code) != 3 || n < 100 {
		return nil, fmt.Errorf("%w: status code %q", ErrMalformed, code)
	}
	return &Response{Proto: proto, StatusCode: n, Reason: reason}, nil
}

// bodyReader decides how long the body is. These are the message-length
// rules of RFC 9112 section 6.3, in priority order.
func bodyReader(r *bufio.Reader, method string, resp *Response) (io.Reader, error) {
	code := resp.StatusCode
	// 1. Some responses never have a body.
	if method == "HEAD" || code < 200 || code == 204 || code == 304 {
		return strings.NewReader(""), nil
	}
	// 2. Transfer-Encoding wins over Content-Length. If chunked is the final
	// coding, the body is self-delimiting.
	if te := resp.Header.Get("Transfer-Encoding"); te != "" {
		codings := strings.Split(te, ",")
		if strings.EqualFold(strings.TrimSpace(codings[len(codings)-1]), "chunked") {
			return &chunkedReader{r: r}, nil
		}
		return r, nil // not chunked: the body runs until the server closes
	}
	// 3. Content-Length gives the exact byte count. Several values are only
	// allowed if they agree.
	if cls := resp.Header.Values("Content-Length"); len(cls) > 0 {
		n, err := strconv.ParseInt(cls[0], 10, 64)
		if err != nil || n < 0 {
			return nil, fmt.Errorf("%w: Content-Length %q", ErrMalformed, cls[0])
		}
		for _, v := range cls[1:] {
			if v != cls[0] {
				return nil, fmt.Errorf("%w: conflicting Content-Length values", ErrMalformed)
			}
		}
		return &exactReader{r: io.LimitReader(r, n), left: n}, nil
	}
	// 4. Otherwise the body is everything until the connection closes.
	return r, nil
}

// readLine reads one line and strips the trailing CRLF (or a bare LF,
// which RFC 9112 allows recipients to accept).
func readLine(r *bufio.Reader) (string, error) {
	var b []byte
	for {
		chunk, err := r.ReadSlice('\n')
		b = append(b, chunk...)
		if len(b) > MaxLineLength {
			return "", ErrLineTooLong
		}
		if err == bufio.ErrBufferFull {
			continue
		}
		if err == io.EOF && len(b) > 0 {
			return "", io.ErrUnexpectedEOF
		}
		if err != nil {
			return "", err
		}
		break
	}
	b = b[:len(b)-1]
	if n := len(b); n > 0 && b[n-1] == '\r' {
		b = b[:n-1]
	}
	return string(b), nil
}

// readHeader reads header field lines up to and including the empty line
// that ends the header section.
func readHeader(r *bufio.Reader) (Header, error) {
	var h Header
	for {
		line, err := readLine(r)
		if err != nil {
			return nil, err
		}
		if line == "" {
			return h, nil
		}
		if len(h) == MaxHeaderCount {
			return nil, ErrTooManyHeaders
		}
		if line[0] == ' ' || line[0] == '\t' {
			// Obsolete line folding: a source of parser disagreements, so reject it.
			return nil, fmt.Errorf("%w: folded header line", ErrMalformed)
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok || name == "" || strings.ContainsAny(name, " \t") {
			return nil, fmt.Errorf("%w: header line %q", ErrMalformed, line)
		}
		h = append(h, Field{Name: name, Value: strings.Trim(value, " \t")})
	}
}

// exactReader reports io.ErrUnexpectedEOF if the connection ends before
// Content-Length bytes have arrived, instead of a silent short body.
type exactReader struct {
	r    io.Reader
	left int64
}

func (e *exactReader) Read(p []byte) (int, error) {
	n, err := e.r.Read(p)
	e.left -= int64(n)
	if err == io.EOF && e.left > 0 {
		err = io.ErrUnexpectedEOF
	}
	return n, err
}

// chunkedReader decodes "Transfer-Encoding: chunked":
//
//	<size in hex>[;extensions]\r\n
//	<size bytes of data>\r\n
//	... repeated ...
//	0\r\n
//	[trailer fields]\r\n
//	\r\n
type chunkedReader struct {
	r       *bufio.Reader
	left    uint64 // data bytes left in the current chunk
	started bool   // at least one chunk header has been read
	err     error  // sticky: once set, every Read returns it
}

func (c *chunkedReader) Read(p []byte) (int, error) {
	if c.err != nil {
		return 0, c.err
	}
	if c.left == 0 {
		if err := c.nextChunk(); err != nil {
			c.err = err
			return 0, err
		}
	}
	if uint64(len(p)) > c.left {
		p = p[:c.left]
	}
	n, err := c.r.Read(p)
	c.left -= uint64(n)
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	c.err = err
	return n, err
}

// nextChunk consumes the CRLF after the previous chunk's data, then the
// next chunk-size line. At the last (zero-size) chunk it reads the
// trailer section and returns io.EOF.
func (c *chunkedReader) nextChunk() error {
	if c.started {
		line, err := readLine(c.r)
		if err != nil {
			return err
		}
		if line != "" {
			return fmt.Errorf("%w: missing CRLF after chunk data", ErrMalformed)
		}
	}
	c.started = true

	line, err := readLine(c.r)
	if err != nil {
		return err
	}
	sizeText, _, _ := strings.Cut(line, ";") // chunk extensions are ignored
	size, err := strconv.ParseUint(strings.TrimSpace(sizeText), 16, 63)
	if err != nil {
		return fmt.Errorf("%w: chunk size %q", ErrMalformed, line)
	}
	if size == 0 {
		if _, err := readHeader(c.r); err != nil { // trailers
			return err
		}
		return io.EOF
	}
	c.left = size
	return nil
}
