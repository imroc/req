package http2

import "testing"

func TestPriorityParamIsZero(t *testing.T) {
	var p PriorityParam
	if !p.IsZero() {
		t.Fatal("zero PriorityParam should be zero")
	}
	p = PriorityParam{StreamDep: 1}
	if p.IsZero() {
		t.Fatal("PriorityParam with StreamDep=1 should not be zero")
	}
}

func TestSettingIDString(t *testing.T) {
	tests := []struct {
		id       SettingID
		expected string
	}{
		{SettingHeaderTableSize, "HEADER_TABLE_SIZE"},
		{SettingEnablePush, "ENABLE_PUSH"},
		{SettingMaxConcurrentStreams, "MAX_CONCURRENT_STREAMS"},
		{SettingInitialWindowSize, "INITIAL_WINDOW_SIZE"},
		{SettingMaxFrameSize, "MAX_FRAME_SIZE"},
		{SettingMaxHeaderListSize, "MAX_HEADER_LIST_SIZE"},
		{SettingID(0x99), "UNKNOWN_SETTING_153"},
	}
	for _, tt := range tests {
		if got := tt.id.String(); got != tt.expected {
			t.Errorf("SettingID(%d).String() = %q, want %q", uint16(tt.id), got, tt.expected)
		}
	}
}

func TestSettingString(t *testing.T) {
	s := Setting{ID: SettingHeaderTableSize, Val: 4096}
	got := s.String()
	expected := "[HEADER_TABLE_SIZE = 4096]"
	if got != expected {
		t.Errorf("Setting.String() = %q, want %q", got, expected)
	}
}

func TestPriorityFrame(t *testing.T) {
	pf := PriorityFrame{
		StreamID: 1,
		PriorityParam: PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    16,
		},
	}
	if pf.StreamID != 1 {
		t.Fatalf("StreamID = %d, want 1", pf.StreamID)
	}
	if pf.PriorityParam.Weight != 16 {
		t.Fatalf("Weight = %d, want 16", pf.PriorityParam.Weight)
	}
	if pf.PriorityParam.IsZero() {
		t.Fatal("PriorityParam should not be zero with Weight=16")
	}
}
