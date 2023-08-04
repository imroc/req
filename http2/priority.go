package http2

// PriorityParam are the stream prioritzation parameters.
type PriorityParam struct {
	// StreamDep is a 31-bit stream identifier for the
	// stream that this stream depends on. Zero means no
	// dependency.
	StreamDep uint32

	// Exclusive is whether the dependency is exclusive.
	Exclusive bool

	// Weight is the stream's zero-indexed weight. It should be
	// set together with StreamDep, or neither should be set. Per
	// the spec, "Add one to the value to obtain a weight between
	// 1 and 256."
	Weight uint8
}

func (p PriorityParam) IsZero() bool {
	return p == PriorityParam{}
}

// PriorityFrame represents a http priority frame.
type PriorityFrame struct {
	StreamID      uint32
	PriorityParam PriorityParam
}
