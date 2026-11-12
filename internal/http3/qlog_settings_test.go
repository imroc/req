package http3

import (
	"bytes"
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/imroc/req/v3/internal/testcert"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3/qlog"
	"github.com/quic-go/quic-go/qlogwriter"
	"github.com/quic-go/quic-go/quicvarint"
)

type qlogFrameRecord struct {
	frame qlog.Frame
	raw   qlog.RawInfo
}

type qlogFrameRecorder struct {
	frames chan qlogFrameRecord
}

func (r *qlogFrameRecorder) AddProducer() qlogwriter.Recorder { return r }
func (r *qlogFrameRecorder) Close() error                     { return nil }

func (r *qlogFrameRecorder) RecordEvent(event qlogwriter.Event) {
	switch event := event.(type) {
	case qlog.FrameCreated:
		r.frames <- qlogFrameRecord{frame: event.Frame, raw: event.Raw}
	case qlog.FrameParsed:
		r.frames <- qlogFrameRecord{frame: event.Frame, raw: event.Raw}
	}
}

func TestSettingsFrameQlogValues(t *testing.T) {
	for _, tt := range []struct {
		name                string
		setting             uint64
		value               uint64
		wantDatagram        int64
		wantExtendedConnect int64
	}{
		{name: "absent", wantDatagram: -1, wantExtendedConnect: -1},
		{name: "datagram disabled", setting: settingDatagram, value: 0, wantDatagram: 0, wantExtendedConnect: -1},
		{name: "datagram enabled", setting: settingDatagram, value: 1, wantDatagram: 1, wantExtendedConnect: -1},
		{name: "extended connect disabled", setting: settingExtendedConnect, value: 0, wantDatagram: -1, wantExtendedConnect: 0},
		{name: "extended connect enabled", setting: settingExtendedConnect, value: 1, wantDatagram: -1, wantExtendedConnect: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var payload []byte
			if tt.setting != 0 {
				payload = quicvarint.Append(payload, tt.setting)
				payload = quicvarint.Append(payload, tt.value)
			}
			recorder := &qlogFrameRecorder{frames: make(chan qlogFrameRecord, 1)}
			r := &countingByteReader{Reader: quicvarint.NewReader(bytes.NewReader(payload))}
			if _, err := parseSettingsFrame(r, uint64(len(payload)), 0, recorder); err != nil {
				t.Fatal(err)
			}
			select {
			case record := <-recorder.frames:
				frame, ok := record.frame.Frame.(qlog.SettingsFrame)
				if !ok {
					t.Fatalf("expected SETTINGS frame, got %T", record.frame.Frame)
				}
				if frame.MaxFieldSectionSize != -1 || frame.Datagram != tt.wantDatagram || frame.ExtendedConnect != tt.wantExtendedConnect {
					t.Fatalf("unexpected logged settings: %+v", frame)
				}
			default:
				t.Fatal("SETTINGS frame was not logged")
			}
		})
	}
}

func TestQlogInvalidHeadersFrameLength(t *testing.T) {
	for _, length := range []uint64{0, 7} {
		recorder := &qlogFrameRecorder{frames: make(chan qlogFrameRecord, 1)}
		maybeQlogInvalidHeadersFrame(recorder, 0, length)
		select {
		case record := <-recorder.frames:
			if record.raw.Length != qlog.OmitLength || record.raw.PayloadLength != int(length) {
				t.Fatalf("unexpected logged lengths: %+v", record.raw)
			}
		default:
			t.Fatal("invalid HEADERS frame was not logged")
		}
	}
}

func TestClientConnQlogSettings(t *testing.T) {
	cert, err := tls.X509KeyPair(testcert.LocalhostCert, testcert.LocalhostKey)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name            string
		enableDatagrams bool
		wantDatagram    int64
	}{
		{name: "datagram absent", wantDatagram: -1},
		{name: "datagram enabled", enableDatagrams: true, wantDatagram: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			listener, err := quic.ListenAddr("127.0.0.1:0", &tls.Config{
				Certificates: []tls.Certificate{cert},
				NextProtos:   []string{"h3"},
			}, &quic.Config{EnableDatagrams: true})
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			recorder := &qlogFrameRecorder{frames: make(chan qlogFrameRecord, 1)}
			conn, err := quic.DialAddr(ctx, listener.Addr().String(), &tls.Config{
				InsecureSkipVerify: true,
				NextProtos:         []string{"h3"},
			}, &quic.Config{
				EnableDatagrams: true,
				Tracer: func(context.Context, bool, quic.ConnectionID) qlogwriter.Trace {
					return recorder
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			defer conn.CloseWithError(0, "")
			serverConn, err := listener.Accept(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer serverConn.CloseWithError(0, "")

			tr := &Transport{EnableDatagrams: tt.enableDatagrams, MaxResponseHeaderBytes: 1024}
			tr.NewClientConn(conn)
			select {
			case record := <-recorder.frames:
				frame, ok := record.frame.Frame.(qlog.SettingsFrame)
				if !ok {
					t.Fatalf("expected SETTINGS frame, got %T", record.frame.Frame)
				}
				if frame.MaxFieldSectionSize != 1024 || frame.Datagram != tt.wantDatagram || frame.ExtendedConnect != -1 {
					t.Fatalf("unexpected logged settings: %+v", frame)
				}
				if record.raw.Length <= 0 || record.raw.PayloadLength != qlog.OmitLength {
					t.Fatalf("unexpected logged lengths: %+v", record.raw)
				}
			case <-ctx.Done():
				t.Fatal("SETTINGS frame was not logged:", ctx.Err())
			}
		})
	}
}
