package http3

import (
	"errors"
	"fmt"
	"github.com/quic-go/qpack"
	"golang.org/x/net/http/httpguts"
	"net/http"
	"strconv"
	"strings"
)

type Header struct {
	// Pseudo header fields defined in RFC 9114
	Path      string
	Method    string
	Authority string
	Scheme    string
	Status    string
	// for Extended connect
	Protocol string
	// parsed and deduplicated
	ContentLength int64
	// all non-pseudo headers
	Headers http.Header
}

func parseHeaders(headers []qpack.HeaderField, isRequest bool) (Header, error) {
	hdr := Header{Headers: make(http.Header, len(headers))}
	var readFirstRegularHeader, readContentLength bool
	var contentLengthStr string
	for _, h := range headers {
		// field names need to be lowercase, see section 4.2 of RFC 9114
		if strings.ToLower(h.Name) != h.Name {
			return Header{}, fmt.Errorf("header field is not lower-case: %s", h.Name)
		}
		if !httpguts.ValidHeaderFieldValue(h.Value) {
			return Header{}, fmt.Errorf("invalid header field value for %s: %q", h.Name, h.Value)
		}
		if h.IsPseudo() {
			if readFirstRegularHeader {
				// all pseudo headers must appear before regular header fields, see section 4.3 of RFC 9114
				return Header{}, fmt.Errorf("received pseudo header %s after a regular header field", h.Name)
			}
			var isResponsePseudoHeader bool // pseudo headers are either valid for requests or for responses
			switch h.Name {
			case ":path":
				hdr.Path = h.Value
			case ":method":
				hdr.Method = h.Value
			case ":authority":
				hdr.Authority = h.Value
			case ":protocol":
				hdr.Protocol = h.Value
			case ":scheme":
				hdr.Scheme = h.Value
			case ":status":
				hdr.Status = h.Value
				isResponsePseudoHeader = true
			default:
				return Header{}, fmt.Errorf("unknown pseudo header: %s", h.Name)
			}
			if isRequest && isResponsePseudoHeader {
				return Header{}, fmt.Errorf("invalid request pseudo header: %s", h.Name)
			}
			if !isRequest && !isResponsePseudoHeader {
				return Header{}, fmt.Errorf("invalid response pseudo header: %s", h.Name)
			}
		} else {
			if !httpguts.ValidHeaderFieldName(h.Name) {
				return Header{}, fmt.Errorf("invalid header field name: %q", h.Name)
			}
			readFirstRegularHeader = true
			switch h.Name {
			case "content-length":
				// Ignore duplicate Content-Length headers.
				// Fail if the duplicates differ.
				if !readContentLength {
					readContentLength = true
					contentLengthStr = h.Value
				} else if contentLengthStr != h.Value {
					return Header{}, fmt.Errorf("contradicting content lengths (%s and %s)", contentLengthStr, h.Value)
				}
			default:
				hdr.Headers.Add(h.Name, h.Value)
			}
		}
	}
	if len(contentLengthStr) > 0 {
		// use ParseUint instead of ParseInt, so that parsing fails on negative values
		cl, err := strconv.ParseUint(contentLengthStr, 10, 63)
		if err != nil {
			return Header{}, fmt.Errorf("invalid content length: %w", err)
		}
		hdr.Headers.Set("Content-Length", contentLengthStr)
		hdr.ContentLength = int64(cl)
	}
	return hdr, nil
}

func hostnameFromRequest(req *http.Request) string {
	if req.URL != nil {
		return req.URL.Host
	}
	return ""
}

func responseFromHeaders(headerFields []qpack.HeaderField) (*http.Response, error) {
	hdr, err := parseHeaders(headerFields, false)
	if err != nil {
		return nil, err
	}
	if hdr.Status == "" {
		return nil, errors.New("missing status field")
	}
	rsp := &http.Response{
		Proto:         "HTTP/3.0",
		ProtoMajor:    3,
		Header:        hdr.Headers,
		ContentLength: hdr.ContentLength,
	}
	status, err := strconv.Atoi(hdr.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %w", err)
	}
	rsp.StatusCode = status
	rsp.Status = hdr.Status + " " + http.StatusText(status)
	return rsp, nil
}
