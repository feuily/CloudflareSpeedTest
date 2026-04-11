package task

import (
	"net/http"
	"sync"
	"time"
)

const (
	defaultHTTPQPSMin    = 10
	defaultHTTPQPSMax    = 30
	defaultHTTPQPSFactor = 4
)

var RequestQPS int

var globalRequestLimiter = &requestLimiter{}

type requestLimiter struct {
	mu      sync.RWMutex
	tokenCh chan struct{}
	stopCh  chan struct{}
}

func configureRequestLimiter() {
	qps := resolveRequestQPS()
	globalRequestLimiter.configure(qps)
}

func resolveRequestQPS() int {
	if RequestQPS > 0 {
		return RequestQPS
	}
	if RequestQPS < 0 {
		return 0
	}
	if !Httping {
		return 0
	}
	qps := Routines / defaultHTTPQPSFactor
	if qps < defaultHTTPQPSMin {
		qps = defaultHTTPQPSMin
	}
	if qps > defaultHTTPQPSMax {
		qps = defaultHTTPQPSMax
	}
	return qps
}

func (r *requestLimiter) configure(qps int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stopCh != nil {
		close(r.stopCh)
		r.stopCh = nil
	}
	r.tokenCh = nil
	if qps <= 0 {
		return
	}

	interval := time.Second / time.Duration(qps)
	if interval <= 0 {
		interval = time.Nanosecond
	}

	tokenCh := make(chan struct{}, 1)
	tokenCh <- struct{}{}
	stopCh := make(chan struct{})
	go func(tokens chan struct{}, stop chan struct{}, tickerInterval time.Duration) {
		ticker := time.NewTicker(tickerInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				select {
				case tokens <- struct{}{}:
				default:
				}
			case <-stop:
				return
			}
		}
	}(tokenCh, stopCh, interval)

	r.tokenCh = tokenCh
	r.stopCh = stopCh
}

func (r *requestLimiter) wait() {
	r.mu.RLock()
	tokenCh := r.tokenCh
	r.mu.RUnlock()
	if tokenCh == nil {
		return
	}
	<-tokenCh
}

type rateLimitedTransport struct {
	base http.RoundTripper
}

func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	globalRequestLimiter.wait()
	return t.base.RoundTrip(req)
}

func newRateLimitedTransport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &rateLimitedTransport{base: base}
}

func waitForRequestSlot() {
	globalRequestLimiter.wait()
}