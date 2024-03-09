package connection

import (
	"sync"
	"time"
)

type proxyPool struct {
	mu      sync.Mutex
	proxies map[string]*proxyPair
}

// registerMobileConn register proxyPair if not existing in pool and returns it.
func (pp *proxyPool) getOrCreateProxyPair(id string) *proxyPair {
	pp.mu.Lock()
	defer pp.mu.Unlock()
	v, ok := pp.proxies[id]
	if !ok {
		v = initProxyPair()
	}
	pp.proxies[id] = v
	return v
}

func (pp *proxyPool) deleteExpiresPairs() {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	for key, pair := range pp.proxies {
		if time.Now().After(pair.expiresAt) {
			delete(pp.proxies, key)
		}
	}
}

type proxyPair struct {
	toMobileDataCh    chan []byte
	toExtensionDataCh chan []byte
	expiresAt         time.Time
}

// initProxyPair returns proxyPair and runs loop responsible for proxing data.
func initProxyPair() *proxyPair {
	const proxyTimeout = 3 * time.Minute
	return &proxyPair{
		toMobileDataCh:    make(chan []byte),
		toExtensionDataCh: make(chan []byte),
		expiresAt:         time.Now().Add(proxyTimeout),
	}
}
