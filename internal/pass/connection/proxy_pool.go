package connection

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/twofas/2fas-server/internal/common/logging"
)

// Proxy between Browser Extension and Mobile.
type Proxy struct {
	proxyPool *proxyPool
	idLabel   string
}

func NewProxy(idLabel string) *Proxy {
	proxyPool := &proxyPool{proxies: map[string]*proxyPair{}}
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for {
			<-ticker.C
			proxyPool.deleteExpiresPairs()
		}
	}()
	return &Proxy{
		proxyPool: proxyPool,
		idLabel:   idLabel,
	}
}

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

func (p *Proxy) ServeExtensionProxyToMobileWS(w http.ResponseWriter, r *http.Request, id string) error {
	log := logging.WithField(p.idLabel, id)
	conn, err := Upgrade(w, r)
	if err != nil {
		return fmt.Errorf("failed to upgrade connection: %w", err)
	}

	log.Infof("Starting ServeExtensionProxyToMobileWS")

	proxyPair := p.proxyPool.getOrCreateProxyPair(id)
	StartProxy(conn, proxyPair.toMobileDataCh, proxyPair.toExtensionDataCh)
	return nil
}

func (p *Proxy) ServeMobileProxyToExtensionWS(w http.ResponseWriter, r *http.Request, id string) error {
	conn, err := Upgrade(w, r)
	if err != nil {
		return fmt.Errorf("failed to upgrade connection: %w", err)
	}

	logging.Infof("Starting ServeMobileProxyToExtensionWS for dev: %v", id)
	proxyPair := p.proxyPool.getOrCreateProxyPair(id)

	StartProxy(conn, proxyPair.toExtensionDataCh, proxyPair.toMobileDataCh)

	return nil
}
