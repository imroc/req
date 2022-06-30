package utils

import "github.com/imroc/req/v3/internal/protocol"

// PacketInterval is an interval from one PacketNumber to the other
type PacketInterval struct {
	Start protocol.PacketNumber
	End   protocol.PacketNumber
}
