// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package req

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"math"
	"net/textproto"
	"sync"

	"github.com/imroc/req/v3/internal/dump"
)

func isASCIILetter(b byte) bool {
	b |= 0x20 // make lower case
	return 'a' <= b && b <= 'z'
}

// TODO: This should be a distinguishable error (ErrMessageTooLarge)
// to allow mime/multipart to detect it.
var errMessageTooLarge = errors.New("message too large")

// A textprotoReader implements convenience methods for reading requests
// or responses from a text protocol network connection.
type textprotoReader struct {
	R        *bufio.Reader
	buf      []byte // a reusable buffer for readContinuedLineSlice
	readLine func() (line []byte, isPrefix bool, err error)
}

// NewReader returns a new textprotoReader reading from r.
//
// To avoid denial of service attacks, the provided bufio.Reader
// should be reading from an io.LimitReader or similar textprotoReader to bound
// the size of responses.
func newTextprotoReader(r *bufio.Reader, ds dump.Dumpers) *textprotoReader {
	commonHeaderOnce.Do(initCommonHeader)
	t := &textprotoReader{R: r}

	if ds.ShouldDump() {
		t.readLine = func() (line []byte, isPrefix bool, err error) {
			line, err = t.R.ReadSlice('\n')
			if len(line) == 0 {
				if err != nil {
					line = nil
				}
				return
			}
			err = nil
			ds.DumpResponseHeader(line)
			if line[len(line)-1] == '\n' {
				drop := 1
				if len(line) > 1 && line[len(line)-2] == '\r' {
					drop = 2
				}
				line = line[:len(line)-drop]
			}
			return
		}
	} else {
		t.readLine = t.R.ReadLine
	}

	return t
}

// ReadLine reads a single line from r,
// eliding the final \n or \r\n from the returned string.
func (r *textprotoReader) ReadLine() (string, error) {
	line, err := r.readLineSlice(-1)
	return string(line), err
}

// readLineSlice reads a single line from r,
// up to lim bytes long (or unlimited if lim is less than 0),
// eliding the final \r or \r\n from the returned string.
func (r *textprotoReader) readLineSlice(lim int64) ([]byte, error) {
	var line []byte

	for {
		l, more, err := r.readLine()
		if err != nil {
			return nil, err
		}
		if lim >= 0 && int64(len(line))+int64(len(l)) > lim {
			return nil, errMessageTooLarge
		}
		// Avoid the copy if the first call produced a full line.
		if line == nil && !more {
			return l, nil
		}
		line = append(line, l...)
		if !more {
			break
		}
	}
	return line, nil
}

// trim returns s with leading and trailing spaces and tabs removed.
// It does not assume Unicode or UTF-8.
func trim(s []byte) []byte {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	n := len(s)
	for n > i && (s[n-1] == ' ' || s[n-1] == '\t') {
		n--
	}
	return s[i:n]
}

// readContinuedLineSlice reads continued lines from the reader buffer,
// returning a byte slice with all lines. The validateFirstLine function
// is run on the first read line, and if it returns an error then this
// error is returned from readContinuedLineSlice.
// It reads up to lim bytes of data (or unlimited if lim is less than 0).
func (r *textprotoReader) readContinuedLineSlice(lim int64, validateFirstLine func([]byte) error) ([]byte, error) {
	if validateFirstLine == nil {
		return nil, fmt.Errorf("missing validateFirstLine func")
	}

	// Read the first line.
	line, err := r.readLineSlice(lim)
	if err != nil {
		return nil, err
	}
	if len(line) == 0 { // blank line - no continuation
		return line, nil
	}

	if err := validateFirstLine(line); err != nil {
		return nil, err
	}

	// Optimistically assume that we have started to buffer the next line
	// and it starts with an ASCII letter (the next header key), or a blank
	// line, so we can avoid copying that buffered data around in memory
	// and skipping over non-existent whitespace.
	if r.R.Buffered() > 1 {
		peek, _ := r.R.Peek(2)
		if len(peek) > 0 && (isASCIILetter(peek[0]) || peek[0] == '\n') ||
			len(peek) == 2 && peek[0] == '\r' && peek[1] == '\n' {
			return trim(line), nil
		}
	}

	// ReadByte or the next readLineSlice will flush the read buffer;
	// copy the slice into buf.
	r.buf = append(r.buf[:0], trim(line)...)

	if lim < 0 {
		lim = math.MaxInt64
	}
	lim -= int64(len(r.buf))

	// Read continuation lines.
	for r.skipSpace() > 0 {
		r.buf = append(r.buf, ' ')
		if int64(len(r.buf)) >= lim {
			return nil, errMessageTooLarge
		}
		line, err := r.readLineSlice(lim - int64(len(r.buf)))
		if err != nil {
			break
		}
		r.buf = append(r.buf, trim(line)...)
	}
	return r.buf, nil
}

// skipSpace skips R over all spaces and returns the number of bytes skipped.
func (r *textprotoReader) skipSpace() int {
	n := 0
	for {
		c, err := r.R.ReadByte()
		if err != nil {
			// Bufio will keep err until next read.
			break
		}
		if c != ' ' && c != '\t' {
			r.R.UnreadByte()
			break
		}
		n++
	}
	return n
}

// A protocolError describes a protocol violation such
// as an invalid response or a hung-up connection.
type protocolError string

func (p protocolError) Error() string {
	return string(p)
}

var colon = []byte(":")

// ReadMIMEHeader reads a MIME-style header from r.
// The header is a sequence of possibly continued Key: Value lines
// ending in a blank line.
// The returned map m maps [CanonicalMIMEHeaderKey](key) to a
// sequence of values in the same order encountered in the input.
//
// For example, consider this input:
//
//	My-Key: Value 1
//	Long-Key: Even
//	       Longer Value
//	My-Key: Value 2
//
// Given that input, ReadMIMEHeader returns the map:
//
//	map[string][]string{
//		"My-Key": {"Value 1", "Value 2"},
//		"Long-Key": {"Even Longer Value"},
//	}
func (r *textprotoReader) ReadMIMEHeader() (textproto.MIMEHeader, error) {
	return r.readMIMEHeader(math.MaxInt64, math.MaxInt64)
}

// readMIMEHeader is a version of ReadMIMEHeader which takes a limit on the header size.
// It is called by the mime/multipart package.
func (r *textprotoReader) readMIMEHeader(maxMemory, maxHeaders int64) (textproto.MIMEHeader, error) {
	// Avoid lots of small slice allocations later by allocating one
	// large one ahead of time which we'll cut up into smaller
	// slices. If this isn't big enough later, we allocate small ones.
	var strs []string
	hint := r.upcomingHeaderKeys()
	if hint > 0 {
		if hint > 1000 {
			hint = 1000 // set a cap to avoid overallocation
		}
		strs = make([]string, hint)
	}

	m := make(textproto.MIMEHeader, hint)

	// Account for 400 bytes of overhead for the MIMEHeader, plus 200 bytes per entry.
	// Benchmarking map creation as of go1.20, a one-entry MIMEHeader is 416 bytes and large
	// MIMEHeaders average about 200 bytes per entry.
	maxMemory -= 400
	const mapEntryOverhead = 200

	// The first line cannot start with a leading space.
	if buf, err := r.R.Peek(1); err == nil && (buf[0] == ' ' || buf[0] == '\t') {
		const errorLimit = 80 // arbitrary limit on how much of the line we'll quote
		line, err := r.readLineSlice(errorLimit)
		if err != nil {
			return m, err
		}
		return m, protocolError("malformed MIME header initial line: " + string(line))
	}

	for {
		kv, err := r.readContinuedLineSlice(maxMemory, mustHaveFieldNameColon)
		if len(kv) == 0 {
			return m, err
		}

		// Key ends at first colon.
		k, v, ok := bytes.Cut(kv, colon)
		if !ok {
			return m, protocolError("malformed MIME header line: " + string(kv))
		}
		key, ok := canonicalMIMEHeaderKey(k)
		if !ok {
			return m, protocolError("malformed MIME header line: " + string(kv))
		}
		for _, c := range v {
			if !validHeaderValueByte(c) {
				return m, protocolError("malformed MIME header line: " + string(kv))
			}
		}

		maxHeaders--
		if maxHeaders < 0 {
			return nil, errMessageTooLarge
		}

		// Skip initial spaces in value.
		value := string(bytes.TrimLeft(v, " \t"))

		vv := m[key]
		if vv == nil {
			maxMemory -= int64(len(key))
			maxMemory -= mapEntryOverhead
		}
		maxMemory -= int64(len(value))
		if maxMemory < 0 {
			return m, errMessageTooLarge
		}
		if vv == nil && len(strs) > 0 {
			// More than likely this will be a single-element key.
			// Most headers aren't multi-valued.
			// Set the capacity on strs[0] to 1, so any future append
			// won't extend the slice into the other strings.
			vv, strs = strs[:1:1], strs[1:]
			vv[0] = value
			m[key] = vv
		} else {
			m[key] = append(vv, value)
		}

		if err != nil {
			return m, err
		}
	}
}

// mustHaveFieldNameColon ensures that, per RFC 7230, the
// field-name is on a single line, so the first line must
// contain a colon.
func mustHaveFieldNameColon(line []byte) error {
	if bytes.IndexByte(line, ':') < 0 {
		return protocolError(fmt.Sprintf("malformed MIME header: missing colon: %q", line))
	}
	return nil
}

var nl = []byte("\n")

// upcomingHeaderKeys returns an approximation of the number of keys
// that will be in this header. If it gets confused, it returns 0.
func (r *textprotoReader) upcomingHeaderKeys() (n int) {
	// Try to determine the 'hint' size.
	r.R.Peek(1) // force a buffer load if empty
	s := r.R.Buffered()
	if s == 0 {
		return
	}
	peek, _ := r.R.Peek(s)
	for len(peek) > 0 && n < 1000 {
		var line []byte
		line, peek, _ = bytes.Cut(peek, nl)
		if len(line) == 0 || (len(line) == 1 && line[0] == '\r') {
			// Blank line separating headers from the body.
			break
		}
		if line[0] == ' ' || line[0] == '\t' {
			// Folded continuation of the previous line.
			continue
		}
		n++
	}
	return n
}

const toLower = 'a' - 'A'

// validHeaderFieldByte reports whether b is a valid byte in a header
// field name. RFC 7230 says:
//
//	header-field   = field-name ":" OWS field-value OWS
//	field-name     = token
//	tchar = "!" / "#" / "$" / "%" / "&" / "'" / "*" / "+" / "-" / "." /
//	        "^" / "_" / "`" / "|" / "~" / DIGIT / ALPHA
//	token = 1*tchar
func validHeaderFieldByte(b byte) bool {
	return int(b) < len(isTokenTable) && isTokenTable[b]
}

// canonicalMIMEHeaderKey is like CanonicalMIMEHeaderKey but is
// allowed to mutate the provided byte slice before returning the
// string.
//
// For invalid inputs (if a contains spaces or non-token bytes), a
// is unchanged and a string copy is returned.
//
// ok is true if the header key contains only valid characters and spaces.
// ReadMIMEHeader accepts header keys containing spaces, but does not
// canonicalize them.
func canonicalMIMEHeaderKey(a []byte) (_ string, ok bool) {
	if len(a) == 0 {
		return "", false
	}

	// See if a looks like a header key. If not, return it unchanged.
	noCanon := false
	for _, c := range a {
		if validHeaderFieldByte(c) {
			continue
		}
		// Don't canonicalize.
		if c == ' ' {
			// We accept invalid headers with a space before the
			// colon, but must not canonicalize them.
			// See https://go.dev/issue/34540.
			noCanon = true
			continue
		}
		return string(a), false
	}
	if noCanon {
		return string(a), true
	}

	upper := true
	for i, c := range a {
		// Canonicalize: first letter upper case
		// and upper case after each dash.
		// (Host, User-Agent, If-Modified-Since).
		// MIME headers are ASCII only, so no Unicode issues.
		if upper && 'a' <= c && c <= 'z' {
			c -= toLower
		} else if !upper && 'A' <= c && c <= 'Z' {
			c += toLower
		}
		a[i] = c
		upper = c == '-' // for next time
	}
	commonHeaderOnce.Do(initCommonHeader)
	// The compiler recognizes m[string(byteSlice)] as a special
	// case, so a copy of a's bytes into a new string does not
	// happen in this map lookup:
	if v := commonHeader[string(a)]; v != "" {
		return v, true
	}
	return string(a), true
}

// validHeaderValueByte reports whether c is a valid byte in a header
// field value. RFC 7230 says:
//
//	field-content  = field-vchar [ 1*( SP / HTAB ) field-vchar ]
//	field-vchar    = VCHAR / obs-text
//	obs-text       = %x80-FF
//
// RFC 5234 says:
//
//	HTAB           =  %x09
//	SP             =  %x20
//	VCHAR          =  %x21-7E
func validHeaderValueByte(c byte) bool {
	// mask is a 128-bit bitmap with 1s for allowed bytes,
	// so that the byte c can be tested with a shift and an and.
	// If c >= 128, then 1<<c and 1<<(c-64) will both be zero.
	// Since this is the obs-text range, we invert the mask to
	// create a bitmap with 1s for disallowed bytes.
	const mask = 0 |
		(1<<(0x7f-0x21)-1)<<0x21 | // VCHAR: %x21-7E
		1<<0x20 | // SP: %x20
		1<<0x09 // HTAB: %x09
	return ((uint64(1)<<c)&^(mask&(1<<64-1)) |
		(uint64(1)<<(c-64))&^(mask>>64)) == 0
}

// commonHeader interns common header strings.
var commonHeader map[string]string

var commonHeaderOnce sync.Once

func initCommonHeader() {
	commonHeader = make(map[string]string)
	for _, v := range []string{
		"Accept",
		"Accept-Charset",
		"Accept-Encoding",
		"Accept-Language",
		"Accept-Ranges",
		"Cache-Control",
		"Cc",
		"Connection",
		"Content-Id",
		"Content-Language",
		"Content-Length",
		"Content-Transfer-Encoding",
		"Content-Type",
		"Cookie",
		"Date",
		"Dkim-Signature",
		"Etag",
		"Expires",
		"From",
		"Host",
		"If-Modified-Since",
		"If-None-Match",
		"In-Reply-To",
		"Last-Modified",
		"Location",
		"Message-Id",
		"Mime-Version",
		"Pragma",
		"Received",
		"Return-Path",
		"Server",
		"Set-Cookie",
		"Subject",
		"To",
		"User-Agent",
		"Via",
		"X-Forwarded-For",
		"X-Imforwards",
		"X-Powered-By",
	} {
		commonHeader[v] = v
	}
}

// isTokenTable is a copy of net/http/lex.go's isTokenTable.
// See https://httpwg.github.io/specs/rfc7230.html#rule.token.separators
var isTokenTable = [127]bool{
	'!':  true,
	'#':  true,
	'$':  true,
	'%':  true,
	'&':  true,
	'\'': true,
	'*':  true,
	'+':  true,
	'-':  true,
	'.':  true,
	'0':  true,
	'1':  true,
	'2':  true,
	'3':  true,
	'4':  true,
	'5':  true,
	'6':  true,
	'7':  true,
	'8':  true,
	'9':  true,
	'A':  true,
	'B':  true,
	'C':  true,
	'D':  true,
	'E':  true,
	'F':  true,
	'G':  true,
	'H':  true,
	'I':  true,
	'J':  true,
	'K':  true,
	'L':  true,
	'M':  true,
	'N':  true,
	'O':  true,
	'P':  true,
	'Q':  true,
	'R':  true,
	'S':  true,
	'T':  true,
	'U':  true,
	'W':  true,
	'V':  true,
	'X':  true,
	'Y':  true,
	'Z':  true,
	'^':  true,
	'_':  true,
	'`':  true,
	'a':  true,
	'b':  true,
	'c':  true,
	'd':  true,
	'e':  true,
	'f':  true,
	'g':  true,
	'h':  true,
	'i':  true,
	'j':  true,
	'k':  true,
	'l':  true,
	'm':  true,
	'n':  true,
	'o':  true,
	'p':  true,
	'q':  true,
	'r':  true,
	's':  true,
	't':  true,
	'u':  true,
	'v':  true,
	'w':  true,
	'x':  true,
	'y':  true,
	'z':  true,
	'|':  true,
	'~':  true,
}
