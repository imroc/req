package http2

import "testing"

func TestInflowInit(t *testing.T) {
	var f inflow
	f.init(65535)
	if f.avail != 65535 {
		t.Fatalf("avail = %d, want 65535", f.avail)
	}
	if f.unsent != 0 {
		t.Fatalf("unsent = %d, want 0", f.unsent)
	}
}

func TestInflowTake(t *testing.T) {
	var f inflow
	f.init(1000)
	if !f.take(500) {
		t.Fatal("take(500) should succeed with avail=1000")
	}
	if f.avail != 500 {
		t.Fatalf("avail = %d, want 500 after take(500)", f.avail)
	}
	if f.take(501) {
		t.Fatal("take(501) should fail with avail=500")
	}
}

func TestInflowAdd(t *testing.T) {
	var f inflow
	f.init(1000)
	// Small add should buffer (less than inflowMinRefresh=4096)
	add := f.add(100)
	if add != 0 {
		t.Fatalf("small add should return 0, got %d", add)
	}
	if f.unsent != 100 {
		t.Fatalf("unsent = %d, want 100", f.unsent)
	}
	// Large add should trigger window update
	add = f.add(5000)
	if add != 5100 {
		t.Fatalf("large add should return 5100, got %d", add)
	}
	if f.unsent != 0 {
		t.Fatalf("unsent should be 0 after flush, got %d", f.unsent)
	}
}

func TestTakeInflows(t *testing.T) {
	var f1, f2 inflow
	f1.init(1000)
	f2.init(500)
	if !takeInflows(&f1, &f2, 400) {
		t.Fatal("takeInflows(400) should succeed")
	}
	if f1.avail != 600 {
		t.Fatalf("f1.avail = %d, want 600", f1.avail)
	}
	if f2.avail != 100 {
		t.Fatalf("f2.avail = %d, want 100", f2.avail)
	}
	if takeInflows(&f1, &f2, 200) {
		t.Fatal("takeInflows(200) should fail (f2 only has 100)")
	}
}

func TestOutflowAvailable(t *testing.T) {
	var conn outflow
	conn.n = 10000
	var stream outflow
	stream.setConnFlow(&conn)
	stream.n = 5000
	// Limited by stream
	if got := stream.available(); got != 5000 {
		t.Fatalf("available = %d, want 5000", got)
	}
	// When conn is lower, limited by conn
	conn.n = 3000
	if got := stream.available(); got != 3000 {
		t.Fatalf("available = %d, want 3000 (conn limited)", got)
	}
}

func TestOutflowTake(t *testing.T) {
	var conn outflow
	conn.n = 10000
	var stream outflow
	stream.setConnFlow(&conn)
	stream.n = 5000
	stream.take(2000)
	if stream.n != 3000 {
		t.Fatalf("stream.n = %d, want 3000", stream.n)
	}
	if conn.n != 8000 {
		t.Fatalf("conn.n = %d, want 8000", conn.n)
	}
}

func TestOutflowAdd(t *testing.T) {
	var f outflow
	f.n = 1000
	if !f.add(500) {
		t.Fatal("add(500) should succeed")
	}
	if f.n != 1500 {
		t.Fatalf("n = %d, want 1500", f.n)
	}
	// Overflow check
	f.n = 1<<31 - 100
	if f.add(200) {
		t.Fatal("add that overflows should return false")
	}
}
