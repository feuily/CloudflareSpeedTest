package task

import (
	"testing"
	"time"
)

func TestFormatDownloadStatus(t *testing.T) {
	if got := formatDownloadStatus("", 123); got != "" {
		t.Fatalf("expected empty ip to render empty text, got %q", got)
	}

	got := formatDownloadStatus("1.1.1.1", 25*1024*1024)
	want := "1.1.1.1 25.00 MB/s"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestDownloadStageHelpers(t *testing.T) {
	originalMinSpeed := MinSpeed
	originalTestCount := TestCount
	t.Cleanup(func() {
		MinSpeed = originalMinSpeed
		TestCount = originalTestCount
	})

	MinSpeed = 0
	TestCount = 10
	if !shouldUseTwoStageDownload(25) {
		t.Fatalf("expected two-stage download for larger candidate set")
	}
	if shouldUseTwoStageDownload(10) {
		t.Fatalf("expected single-stage download when candidate set is small")
	}

	if got := calculatePreviewTimeout(10 * time.Second); got != 2*time.Second {
		t.Fatalf("expected preview timeout to cap at 2s, got %v", got)
	}
	if got := calculatePreviewTimeout(2 * time.Second); got != 1*time.Second {
		t.Fatalf("expected preview timeout to floor at 1s, got %v", got)
	}

	if got := calculateConfirmCount(25); got != 20 {
		t.Fatalf("expected confirm count to be 20, got %d", got)
	}
	TestCount = 20
	if got := calculateConfirmCount(8); got != 8 {
		t.Fatalf("expected confirm count to respect candidate size, got %d", got)
	}
}
