package task

import "testing"

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