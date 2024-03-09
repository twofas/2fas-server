package connection

import (
	"fmt"
	"net/http"
	"time"

	"github.com/twofas/2fas-server/internal/common/logging"
)

// ProxyServer manages proxy connections between Browser Extension and Mobile.
type ProxyServer struct {
	proxyPool *proxyPool
	idLabel   string
}

func NewProxyServer(idLabel string) *ProxyServer {
	proxyPool := &proxyPool{proxies: map[string]*proxyPair{}}
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for {
			<-ticker.C
			proxyPool.deleteExpiresPairs()
		}
	}()
	return &ProxyServer{
		proxyPool: proxyPool,
		idLabel:   idLabel,
	}
}

func (p *ProxyServer) ServeExtensionProxyToMobileWS(w http.ResponseWriter, r *http.Request, id string) error {
	log := logging.WithField(p.idLabel, id)
	conn, err := Upgrade(w, r)
	if err != nil {
		return fmt.Errorf("failed to upgrade proxy: %w", err)
	}

	log.Infof("Starting ServeExtensionProxyToMobileWS")

	proxyPair := p.proxyPool.getOrCreateProxyPair(id)
	startProxy(conn, proxyPair.toMobileDataCh, proxyPair.toExtensionDataCh)
	return nil
}

func (p *ProxyServer) ServeMobileProxyToExtensionWS(w http.ResponseWriter, r *http.Request, id string) error {
	conn, err := Upgrade(w, r)
	if err != nil {
		return fmt.Errorf("failed to upgrade proxy: %w", err)
	}

	logging.Infof("Starting ServeMobileProxyToExtensionWS for dev: %v", id)
	proxyPair := p.proxyPool.getOrCreateProxyPair(id)

	startProxy(conn, proxyPair.toExtensionDataCh, proxyPair.toMobileDataCh)

	return nil
}
