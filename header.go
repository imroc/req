package req

import (
	"io"
	"net/http"
	"net/textproto"
	"sort"
	"strings"
	"sync"

	"golang.org/x/net/http/httpguts"

	"github.com/imroc/req/v3/internal/header"
)

var headerNewlineToSpace = strings.NewReplacer("\n", " ", "\r", " ")

// stringWriter implements WriteString on a Writer.
type stringWriter struct {
	w io.Writer
}

func (w stringWriter) WriteString(s string) (n int, err error) {
	return w.w.Write([]byte(s))
}

// A headerSorter implements sort.Interface by sorting a []keyValues
// by key. It's used as a pointer, so it can fit in a sort.Interface
// interface value without allocation.
type headerSorter struct {
	kvs []header.KeyValues
}

func (s *headerSorter) Len() int           { return len(s.kvs) }
func (s *headerSorter) Swap(i, j int)      { s.kvs[i], s.kvs[j] = s.kvs[j], s.kvs[i] }
func (s *headerSorter) Less(i, j int) bool { return s.kvs[i].Key < s.kvs[j].Key }

var headerSorterPool = sync.Pool{
	New: func() interface{} { return new(headerSorter) },
}

// get is like Get, but key must already be in CanonicalHeaderKey form.
func headerGet(h http.Header, key string) string {
	if v := h[key]; len(v) > 0 {
		return v[0]
	}
	return ""
}

// has reports whether h has the provided key defined, even if it's
// set to 0-length slice.
func headerHas(h http.Header, key string) bool {
	_, ok := h[key]
	return ok
}

// sortedKeyValues returns h's keys sorted in the returned kvs
// slice. The headerSorter used to sort is also returned, for possible
// return to headerSorterCache.
func headerSortedKeyValues(h http.Header, exclude map[string]bool) (kvs []header.KeyValues, hs *headerSorter) {
	hs = headerSorterPool.Get().(*headerSorter)
	if cap(hs.kvs) < len(h) {
		hs.kvs = make([]header.KeyValues, 0, len(h))
	}
	kvs = hs.kvs[:0]
	for k, vv := range h {
		if !exclude[k] {
			kvs = append(kvs, header.KeyValues{k, vv})
		}
	}
	hs.kvs = kvs
	sort.Sort(hs)
	return kvs, hs
}

func headerWrite(h http.Header, writeHeader func(key string, values ...string) error, sort bool) error {
	return headerWriteSubset(h, nil, writeHeader, sort)
}

func headerWriteSubset(h http.Header, exclude map[string]bool, writeHeader func(key string, values ...string) error, sort bool) error {
	var kvs []header.KeyValues
	var hs *headerSorter
	if sort {
		kvs = make([]header.KeyValues, 0, len(h))
		for k, v := range h {
			if !exclude[k] {
				kvs = append(kvs, header.KeyValues{k, v})
			}
		}
	} else {
		kvs, hs = headerSortedKeyValues(h, exclude)
	}
	for _, kv := range kvs {
		if !httpguts.ValidHeaderFieldName(kv.Key) {
			// This could be an error. In the common case of
			// writing response headers, however, we have no good
			// way to provide the error back to the server
			// handler, so just drop invalid headers instead.
			continue
		}
		for i, v := range kv.Values {
			vv := headerNewlineToSpace.Replace(v)
			vv = textproto.TrimString(v)
			if vv != v {
				kv.Values[i] = vv
			}
		}
		err := writeHeader(kv.Key, kv.Values...)
		if err != nil {
			if hs != nil {
				headerSorterPool.Put(hs)
			}
			return err
		}
	}
	if hs != nil {
		headerSorterPool.Put(hs)
	}
	return nil
}
