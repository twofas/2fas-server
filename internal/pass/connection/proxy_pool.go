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
	toMobileDataCh    *safeChannel
	toExtensionDataCh *safeChannel
	expiresAt         time.Time
}

// initProxyPair returns proxyPair and runs loop responsible for proxing data.
func initProxyPair() *proxyPair {
	const proxyTimeout = 3 * time.Minute
	return &proxyPair{
		toMobileDataCh:    newSafeChannel(),
		toExtensionDataCh: newSafeChannel(),
		expiresAt:         time.Now().Add(proxyTimeout),
	}
}

type safeChannel struct {
	channel chan []byte
	mu      *sync.Mutex
}

func newSafeChannel() *safeChannel {
	return &safeChannel{
		channel: make(chan []byte),
		mu:      &sync.Mutex{},
	}
}

func (sc *safeChannel) write(data []byte) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.channel == nil {
		return
	}

	sc.channel <- data
}

func (sc *safeChannel) close() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.channel == nil {
		return
	}

	close(sc.channel)
	sc.channel = nil
}
