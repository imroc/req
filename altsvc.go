package req

import (
	"github.com/imroc/req/v3/pkg/altsvc"
	"net/http"
	"sync"
	"time"
)

type pendingAltSvc struct {
	CurrentIndex int
	Entries      []*altsvc.AltSvc
	Mu           sync.Mutex
	LastTime     time.Time
	Transport    http.RoundTripper
}
