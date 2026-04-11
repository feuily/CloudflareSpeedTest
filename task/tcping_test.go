package task

import "testing"

func TestCheckPingDefaultCapsHTTPingRoutines(t *testing.T) {
	originalHttping := Httping
	originalRoutines := Routines
	t.Cleanup(func() {
		Httping = originalHttping
		Routines = originalRoutines
	})

	Httping = true
	Routines = 500
	checkPingDefault()

	if Routines != httpingRoutineCap {
		t.Fatalf("expected HTTPing routines to be capped at %d, got %d", httpingRoutineCap, Routines)
	}
}