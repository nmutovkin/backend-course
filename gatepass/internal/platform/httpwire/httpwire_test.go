package httpwire

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestReadResponse(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		raw      string
		wantCode int
		wantBody string
		wantErr  error // checked with errors.Is against ReadResponse or body read
	}{
		{
			name:     "content-length",
			raw:      "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\nhelloEXTRA",
			wantCode: 200, wantBody: "hello",
		},
		{
			name: "chunked with extension and trailer",
			raw: "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n" +
				"4;note=x\r\nWiki\r\n" +
				"7\r\npedia i\r\n" +
				"B\r\nn chunks.\r\n\r\n" +
				"0\r\nServer-Timing: db;dur=3\r\n\r\n",
			wantCode: 200, wantBody: "Wikipedia in chunks.\r\n",
		},
		{
			name:     "read until close",
			raw:      "HTTP/1.0 200 OK\r\n\r\nall of it",
			wantCode: 200, wantBody: "all of it",
		},
		{
			name:     "bare LF line endings are accepted",
			raw:      "HTTP/1.1 200 OK\nContent-Length: 2\n\nok",
			wantCode: 200, wantBody: "ok",
		},
		{
			name:     "HEAD has no body",
			method:   "HEAD",
			raw:      "HTTP/1.1 200 OK\r\nContent-Length: 1000\r\n\r\n",
			wantCode: 200, wantBody: "",
		},
		{
			name:     "204 has no body",
			raw:      "HTTP/1.1 204 No Content\r\n\r\n",
			wantCode: 204, wantBody: "",
		},
		{
			name:     "empty reason phrase",
			raw:      "HTTP/1.1 200 \r\nContent-Length: 0\r\n\r\n",
			wantCode: 200, wantBody: "",
		},
		{
			name:    "malformed status line",
			raw:     "HTTP/1.1 OK\r\n\r\n",
			wantErr: ErrMalformed,
		},
		{
			name:    "conflicting content-length",
			raw:     "HTTP/1.1 200 OK\r\nContent-Length: 5\r\nContent-Length: 6\r\n\r\nhello",
			wantErr: ErrMalformed,
		},
		{
			name:    "folded header",
			raw:     "HTTP/1.1 200 OK\r\nX-A: 1\r\n  continued\r\n\r\n",
			wantErr: ErrMalformed,
		},
		{
			name:     "bad chunk size",
			raw:      "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\nzz\r\n",
			wantCode: 200, wantErr: ErrMalformed,
		},
		{
			name:     "body shorter than content-length",
			raw:      "HTTP/1.1 200 OK\r\nContent-Length: 10\r\n\r\nshort",
			wantCode: 200, wantErr: io.ErrUnexpectedEOF,
		},
		{
			name:     "connection drops mid-chunk",
			raw:      "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\nA\r\nabc",
			wantCode: 200, wantErr: io.ErrUnexpectedEOF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := tt.method
			if method == "" {
				method = "GET"
			}
			resp, err := ReadResponse(bufio.NewReader(strings.NewReader(tt.raw)), method)
			if err == nil {
				if resp.StatusCode != tt.wantCode {
					t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantCode)
				}
				var body []byte
				body, err = io.ReadAll(resp.Body)
				if err == nil && string(body) != tt.wantBody {
					t.Errorf("body = %q, want %q", body, tt.wantBody)
				}
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestHeaderLookupIsCaseInsensitive(t *testing.T) {
	h := Header{{"Set-Cookie", "a=1"}, {"content-type", "text/plain"}, {"set-cookie", "b=2"}}
	if got := h.Get("Content-Type"); got != "text/plain" {
		t.Errorf("Get = %q", got)
	}
	if got := h.Values("SET-COOKIE"); len(got) != 2 || got[0] != "a=1" || got[1] != "b=2" {
		t.Errorf("Values = %q", got)
	}
}

func TestLineTooLong(t *testing.T) {
	raw := "HTTP/1.1 200 OK\r\nX-Big: " + strings.Repeat("a", MaxLineLength) + "\r\n\r\n"
	_, err := ReadResponse(bufio.NewReader(strings.NewReader(raw)), "GET")
	if !errors.Is(err, ErrLineTooLong) {
		t.Fatalf("err = %v, want ErrLineTooLong", err)
	}
}
