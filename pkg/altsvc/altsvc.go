package altsvc

import (
	"sync"
	"time"
)

type AltSvcJar struct {
	entries map[string]*AltSvc
	mu      sync.Mutex
}

func NewAltSvcJar() *AltSvcJar {
	return &AltSvcJar{
		entries: make(map[string]*AltSvc),
	}
}

func (j *AltSvcJar) GetAltSvc(addr string) *AltSvc {
	if addr == "" {
		return nil
	}
	as, ok := j.entries[addr]
	if !ok {
		return nil
	}
	now := time.Now()
	j.mu.Lock()
	defer j.mu.Unlock()
	if as.Expire.Before(now) { // expired
		delete(j.entries, addr)
		return nil
	}
	return as
}

func (j *AltSvcJar) SetAltSvc(addr string, as *AltSvc) {
	if addr == "" {
		return
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	j.entries[addr] = as
}

type AltSvc struct {
	Protocol string
	Host     string
	Port     string
	Expire   time.Time
}
