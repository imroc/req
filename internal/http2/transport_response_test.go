package http2

import (
	"github.com/imroc/req/v3/internal/transport"
	"golang.org/x/net/http2/hpack"
	"testing"
)

func TestHandleResponseFramingSecurity(t *testing.T) {
	makeMetaFrame := func(fields []hpack.HeaderField, endStream bool) *MetaHeadersFrame {
		flags := Flags(0)
		if endStream {
			flags |= FlagHeadersEndStream
		}
		return &MetaHeadersFrame{
			HeadersFrame: &HeadersFrame{
				FrameHeader: FrameHeader{Flags: flags},
			},
			Fields: fields,
		}
	}

	t.Run("strips connection-specific headers", func(t *testing.T) {
		fields := []hpack.HeaderField{
			{Name: ":status", Value: "200"},
			{Name: "content-type", Value: "text/plain"},
			{Name: "connection", Value: "keep-alive"},
			{Name: "keep-alive", Value: "timeout=5"},
			{Name: "proxy-connection", Value: "keep-alive"},
			{Name: "transfer-encoding", Value: "chunked"},
			{Name: "upgrade", Value: "websocket"},
		}

		rl := &clientConnReadLoop{}
		cs := &clientStream{}
		f := makeMetaFrame(fields, true)

		res, err := rl.handleResponse(cs, f)
		if err != nil {
			t.Fatalf("handleResponse failed: %v", err)
		}

		if got := res.Header.Get("Content-Type"); got != "text/plain" {
			t.Errorf("Content-Type = %q, want text/plain", got)
		}

		for _, bad := range connHeaders {
			if v, ok := res.Header[bad]; ok {
				t.Errorf("header %q not deleted, got: %v", bad, v)
			}
		}
	})

	t.Run("strict content-length handling", func(t *testing.T) {
		tests := []struct {
			name        string
			clValues    []string
			wantLen     int64
			wantPresent bool
		}{
			{
				name:        "valid single",
				clValues:    []string{"123"},
				wantLen:     123,
				wantPresent: true,
			},
			{
				name:        "identical duplicate values collapsed",
				clValues:    []string{"42", "42", "42"},
				wantLen:     42,
				wantPresent: true,
			},
			{
				name:        "conflicting duplicate values stripped",
				clValues:    []string{"42", "100"},
				wantLen:     0, // f.StreamEnded() sets 0 when ContentLength < 0
				wantPresent: false,
			},
			{
				name:        "invalid non-numeric value stripped",
				clValues:    []string{"invalid"},
				wantLen:     0,
				wantPresent: false,
			},
			{
				name:        "plus signed value stripped",
				clValues:    []string{"+42"},
				wantLen:     0,
				wantPresent: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fields := []hpack.HeaderField{
					{Name: ":status", Value: "200"},
				}
				for _, v := range tt.clValues {
					fields = append(fields, hpack.HeaderField{Name: "content-length", Value: v})
				}

				rl := &clientConnReadLoop{}
				cs := &clientStream{}
				f := makeMetaFrame(fields, true)

				res, err := rl.handleResponse(cs, f)
				if err != nil {
					t.Fatalf("handleResponse failed: %v", err)
				}

				if res.ContentLength != tt.wantLen {
					t.Errorf("ContentLength = %d, want %d", res.ContentLength, tt.wantLen)
				}

				clHeader, ok := res.Header["Content-Length"]
				if ok != tt.wantPresent {
					t.Errorf("Header[Content-Length] present = %v, want %v", ok, tt.wantPresent)
				}
				if tt.wantPresent {
					if len(clHeader) != 1 || clHeader[0] != tt.clValues[0] {
						t.Errorf("Header[Content-Length] = %v, want [%s]", clHeader, tt.clValues[0])
					}
				}
			})
		}
	})

	t.Run("preserves ContentLength = -1 when stream not ended", func(t *testing.T) {
		fields := []hpack.HeaderField{
			{Name: ":status", Value: "200"},
			{Name: "content-length", Value: "bad"},
		}

		rl := &clientConnReadLoop{}
		cs := &clientStream{
			cc: &ClientConn{
				t: &Transport{
					Options: &transport.Options{},
				},
			},
		}
		f := makeMetaFrame(fields, false) // stream not ended

		res, err := rl.handleResponse(cs, f)
		if err != nil {
			t.Fatalf("handleResponse failed: %v", err)
		}

		if res.ContentLength != -1 {
			t.Errorf("ContentLength = %d, want -1 for malformed Content-Length on open stream", res.ContentLength)
		}
		if _, ok := res.Header["Content-Length"]; ok {
			t.Errorf("Header[Content-Length] unexpectedly present")
		}
	})
}
