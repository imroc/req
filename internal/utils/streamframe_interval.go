package utils

import "github.com/imroc/req/v3/internal/protocol"

// ByteInterval is an interval from one ByteCount to the other
type ByteInterval struct {
	Start protocol.ByteCount
	End   protocol.ByteCount
}
