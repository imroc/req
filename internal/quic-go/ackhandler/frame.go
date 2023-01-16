package ackhandler

import "github.com/imroc/req/v3/internal/quic-go/wire"

type Frame struct {
	wire.Frame // nil if the frame has already been acknowledged in another packet
	OnLost     func(wire.Frame)
	OnAcked    func(wire.Frame)
}
