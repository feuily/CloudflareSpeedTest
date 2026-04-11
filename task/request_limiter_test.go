package task

import "testing"

func TestRequestLimiterAdaptiveBackoff(t *testing.T) {
	limiter := &requestLimiter{}
	limiter.configure(20, true)
	t.Cleanup(func() {
		limiter.mu.Lock()
		limiter.stopLocked()
		limiter.mu.Unlock()
	})

	limiter.reportFailure()
	if limiter.qps != 20 {
		t.Fatalf("expected first failure to keep qps unchanged, got %d", limiter.qps)
	}
	limiter.reportFailure()
	if limiter.qps >= 20 {
		t.Fatalf("expected qps to drop after repeated failures, got %d", limiter.qps)
	}

	limiter.mu.Lock()
	limiter.qps = 10
	limiter.baseQPS = 20
	limiter.successStreak = successAdjustThreshold - 1
	limiter.mu.Unlock()
	limiter.reportSuccess()
	if limiter.qps <= 10 {
		t.Fatalf("expected qps to recover after success streak, got %d", limiter.qps)
	}
}

func TestResolveRequestQPS(t *testing.T) {
	originalHttping := Httping
	originalRoutines := Routines
	originalRequestQPS := RequestQPS
	t.Cleanup(func() {
		Httping = originalHttping
		Routines = originalRoutines
		RequestQPS = originalRequestQPS
	})

	RequestQPS = 25
	if got := resolveRequestQPS(); got != 25 {
		t.Fatalf("expected explicit qps to win, got %d", got)
	}

	RequestQPS = -1
	if got := resolveRequestQPS(); got != 0 {
		t.Fatalf("expected negative qps to disable limiter, got %d", got)
	}

	RequestQPS = 0
	Httping = true
	Routines = 200
	if got := resolveRequestQPS(); got != defaultHTTPQPSMax {
		t.Fatalf("expected auto HTTP qps to cap at %d, got %d", defaultHTTPQPSMax, got)
	}

	Routines = 12
	if got := resolveRequestQPS(); got != defaultHTTPQPSMin {
		t.Fatalf("expected auto HTTP qps to floor at %d, got %d", defaultHTTPQPSMin, got)
	}

	Httping = false
	if got := resolveRequestQPS(); got != 0 {
		t.Fatalf("expected non-httping auto qps to be disabled, got %d", got)
	}
}
