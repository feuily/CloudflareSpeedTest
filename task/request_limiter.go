package task

import (
	"net/http"
	"sync"
	"time"

	"github.com/XIU2/CloudflareSpeedTest/utils"
)

const (
	defaultHTTPQPSMin      = 10
	defaultHTTPQPSMax      = 30
	defaultHTTPQPSFactor   = 4
	adaptiveHTTPQPSFloor   = 4
	failureAdjustThreshold = 2
	successAdjustThreshold = 24
	minQPSRecoveryIncrease = 1
)

var RequestQPS int

var globalRequestLimiter = &requestLimiter{}

type requestLimiter struct {
	mu            sync.RWMutex
	tokenCh       chan struct{}
	stopCh        chan struct{}
	qps           int
	baseQPS       int
	autoAdjust    bool
	failureStreak int
	successStreak int
}

func configureRequestLimiter() {
	qps := resolveRequestQPS()
	autoAdjust := RequestQPS == 0 && Httping && qps > 0
	globalRequestLimiter.configure(qps, autoAdjust)
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

func (r *requestLimiter) configure(qps int, autoAdjust bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.stopLocked()
	r.qps = qps
	r.baseQPS = qps
	r.autoAdjust = autoAdjust
	r.failureStreak = 0
	r.successStreak = 0
	if qps <= 0 {
		return
	}
	r.startLocked(qps)
}

func (r *requestLimiter) startLocked(qps int) {
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

func (r *requestLimiter) stopLocked() {
	if r.stopCh != nil {
		close(r.stopCh)
		r.stopCh = nil
	}
	r.tokenCh = nil
}

func (r *requestLimiter) adjustLocked(nextQPS int, reason string) {
	if nextQPS <= 0 || nextQPS == r.qps {
		return
	}
	oldQPS := r.qps
	r.stopLocked()
	r.qps = nextQPS
	r.startLocked(nextQPS)
	if reason != "" {
		utils.Yellow.Printf("[提示] 请求速率自动调整：%d -> %d QPS（%s）\n", oldQPS, nextQPS, reason)
	}
}

func (r *requestLimiter) reportSuccess() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.autoAdjust || r.qps <= 0 {
		return
	}
	r.failureStreak = 0
	r.successStreak++
	if r.successStreak < successAdjustThreshold {
		return
	}
	r.successStreak = 0
	if r.qps >= r.baseQPS {
		return
	}
	nextQPS := r.qps + maxInt(minQPSRecoveryIncrease, r.qps/5)
	if nextQPS > r.baseQPS {
		nextQPS = r.baseQPS
	}
	r.adjustLocked(nextQPS, "连续成功，逐步恢复")
}

func (r *requestLimiter) reportFailure() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.autoAdjust || r.qps <= 0 {
		return
	}
	r.successStreak = 0
	r.failureStreak++
	if r.failureStreak < failureAdjustThreshold {
		return
	}
	r.failureStreak = 0
	if r.qps <= adaptiveHTTPQPSFloor {
		return
	}
	nextQPS := r.qps * 3 / 4
	if nextQPS < adaptiveHTTPQPSFloor {
		nextQPS = adaptiveHTTPQPSFloor
	}
	r.adjustLocked(nextQPS, "连续失败，自动退避")
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func reportRequestSuccess() {
	globalRequestLimiter.reportSuccess()
}

func reportRequestFailure() {
	globalRequestLimiter.reportFailure()
}
