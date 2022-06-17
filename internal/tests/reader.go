package tests

type NeverEnding byte

func (b NeverEnding) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(b)
	}
	return len(p), nil
}
