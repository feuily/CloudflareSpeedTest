package task

import "testing"

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