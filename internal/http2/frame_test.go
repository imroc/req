package http2

import (
	"bytes"
	"testing"

	reqhttp2 "github.com/imroc/req/v3/http2"
)

func TestFrameTypeString(t *testing.T) {
	tests := []struct {
		ft       FrameType
		expected string
	}{
		{FrameData, "DATA"},
		{FrameHeaders, "HEADERS"},
		{FramePriority, "PRIORITY"},
		{FrameRSTStream, "RST_STREAM"},
		{FrameSettings, "SETTINGS"},
		{FramePushPromise, "PUSH_PROMISE"},
		{FramePing, "PING"},
		{FrameGoAway, "GOAWAY"},
		{FrameWindowUpdate, "WINDOW_UPDATE"},
		{FrameContinuation, "CONTINUATION"},
		{FrameType(0xFF), "UNKNOWN_FRAME_TYPE_255"},
	}
	for _, tt := range tests {
		if got := tt.ft.String(); got != tt.expected {
			t.Errorf("FrameType(%d).String() = %q, want %q", uint8(tt.ft), got, tt.expected)
		}
	}
}

func TestFlagsHas(t *testing.T) {
	flags := FlagDataEndStream | FlagDataPadded
	if !flags.Has(FlagDataEndStream) {
		t.Fatal("should have FlagDataEndStream")
	}
	if !flags.Has(FlagDataPadded) {
		t.Fatal("should have FlagDataPadded")
	}
	if flags.Has(FlagHeadersEndHeaders) {
		t.Fatal("should not have FlagHeadersEndHeaders")
	}
}

func TestFrameHeaderString(t *testing.T) {
	fh := FrameHeader{
		Type:     FrameData,
		Flags:    FlagDataEndStream,
		Length:   100,
		StreamID: 1,
	}
	s := fh.String()
	if s == "" {
		t.Fatal("FrameHeader.String() should not be empty")
	}
}

func TestFramerWriteReadData(t *testing.T) {
	var buf bytes.Buffer
	fw := NewFramer(&buf, nil)
	fw.WriteData(1, true, []byte("hello world"))
	
	fr := NewFramer(nil, &buf)
	f, err := fr.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame failed: %v", err)
	}
	df, ok := f.(*DataFrame)
	if !ok {
		t.Fatalf("expected *DataFrame, got %T", f)
	}
	if df.StreamID != 1 {
		t.Fatalf("StreamID = %d, want 1", df.StreamID)
	}
	if !df.StreamEnded() {
		t.Fatal("expected stream to end")
	}
	if string(df.Data()) != "hello world" {
		t.Fatalf("Data = %q, want 'hello world'", string(df.Data()))
	}
}

func TestFramerWriteReadSettings(t *testing.T) {
	var buf bytes.Buffer
	fw := NewFramer(&buf, nil)
	fw.WriteSettings(reqhttp2.Setting{ID: reqhttp2.SettingMaxFrameSize, Val: 16384})

	fr := NewFramer(nil, &buf)
	f, err := fr.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame failed: %v", err)
	}
	sf, ok := f.(*SettingsFrame)
	if !ok {
		t.Fatalf("expected *SettingsFrame, got %T", f)
	}
	if sf.IsAck() {
		t.Fatal("should not be ack")
	}
	if sf.NumSettings() != 1 {
		t.Fatalf("NumSettings = %d, want 1", sf.NumSettings())
	}
	s := sf.Setting(0)
	if s.ID != reqhttp2.SettingMaxFrameSize || s.Val != 16384 {
		t.Fatalf("Setting = %v, want {MaxFrameSize, 16384}", s)
	}
}

func TestFramerWriteReadPing(t *testing.T) {
	var buf bytes.Buffer
	fw := NewFramer(&buf, nil)
	data := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	fw.WritePing(false, data)

	fr := NewFramer(nil, &buf)
	f, err := fr.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame failed: %v", err)
	}
	pf, ok := f.(*PingFrame)
	if !ok {
		t.Fatalf("expected *PingFrame, got %T", f)
	}
	if pf.IsAck() {
		t.Fatal("should not be ack")
	}
	if pf.Data != data {
		t.Fatalf("Data = %v, want %v", pf.Data, data)
	}
}

func TestFramerWriteReadWindowUpdate(t *testing.T) {
	var buf bytes.Buffer
	fw := NewFramer(&buf, nil)
	fw.WriteWindowUpdate(1, 65535)

	fr := NewFramer(nil, &buf)
	f, err := fr.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame failed: %v", err)
	}
	wf, ok := f.(*WindowUpdateFrame)
	if !ok {
		t.Fatalf("expected *WindowUpdateFrame, got %T", f)
	}
	if wf.StreamID != 1 {
		t.Fatalf("StreamID = %d, want 1", wf.StreamID)
	}
	if wf.Increment != 65535 {
		t.Fatalf("Increment = %d, want 65535", wf.Increment)
	}
}

func TestFramerWriteReadGoAway(t *testing.T) {
	var buf bytes.Buffer
	fw := NewFramer(&buf, nil)
	fw.WriteGoAway(0, ErrCodeNo, []byte("shutdown"))

	fr := NewFramer(nil, &buf)
	f, err := fr.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame failed: %v", err)
	}
	gf, ok := f.(*GoAwayFrame)
	if !ok {
		t.Fatalf("expected *GoAwayFrame, got %T", f)
	}
	if gf.LastStreamID != 0 {
		t.Fatalf("LastStreamID = %d, want 0", gf.LastStreamID)
	}
	if gf.ErrCode != ErrCodeNo {
		t.Fatalf("ErrCode = %d, want %d", gf.ErrCode, ErrCodeNo)
	}
	if string(gf.DebugData()) != "shutdown" {
		t.Fatalf("DebugData = %q, want 'shutdown'", string(gf.DebugData()))
	}
}

func TestErrCodeString(t *testing.T) {
	tests := []struct {
		code     ErrCode
		expected string
	}{
		{ErrCodeNo, "NO_ERROR"},
		{ErrCodeProtocol, "PROTOCOL_ERROR"},
		{ErrCodeFlowControl, "FLOW_CONTROL_ERROR"},
		{ErrCodeStreamClosed, "STREAM_CLOSED"},
	}
	for _, tt := range tests {
		if got := tt.code.String(); got != tt.expected {
			t.Errorf("ErrCode(%d).String() = %q, want %q", uint32(tt.code), got, tt.expected)
		}
	}
}
