package utils

import (
	"encoding/csv"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFilterDelayScansAllEntries(t *testing.T) {
	originalMaxDelay := InputMaxDelay
	originalMinDelay := InputMinDelay
	InputMaxDelay = 200 * time.Millisecond
	InputMinDelay = 0
	t.Cleanup(func() {
		InputMaxDelay = originalMaxDelay
		InputMinDelay = originalMinDelay
	})

	data := PingDelaySet{
		{
			PingData: &PingData{IP: &net.IPAddr{IP: net.ParseIP("1.1.1.1")}, Sended: 4, Received: 4, Delay: 300 * time.Millisecond},
		},
		{
			PingData: &PingData{IP: &net.IPAddr{IP: net.ParseIP("1.1.1.2")}, Sended: 4, Received: 3, Delay: 100 * time.Millisecond},
		},
		{
			PingData: &PingData{IP: &net.IPAddr{IP: net.ParseIP("1.1.1.3")}, Sended: 4, Received: 3, Delay: 150 * time.Millisecond},
		},
	}

	filtered := data.FilterDelay()
	if len(filtered) != 2 {
		t.Fatalf("expected 2 entries after delay filtering, got %d", len(filtered))
	}
	if filtered[0].IP.String() != "1.1.1.2" || filtered[1].IP.String() != "1.1.1.3" {
		t.Fatalf("unexpected filtered IP order: %v, %v", filtered[0].IP, filtered[1].IP)
	}
}

func TestExportCsvRespectsPrintNum(t *testing.T) {
	originalOutput := Output
	originalPrintNum := PrintNum
	t.Cleanup(func() {
		Output = originalOutput
		PrintNum = originalPrintNum
	})

	Output = filepath.Join(t.TempDir(), "result.csv")
	PrintNum = 2

	data := []CloudflareIPData{
		{PingData: &PingData{IP: &net.IPAddr{IP: net.ParseIP("1.1.1.1")}, Sended: 4, Received: 4, Delay: 10 * time.Millisecond}, DownloadSpeed: 30 * 1024 * 1024},
		{PingData: &PingData{IP: &net.IPAddr{IP: net.ParseIP("1.1.1.2")}, Sended: 4, Received: 4, Delay: 11 * time.Millisecond}, DownloadSpeed: 20 * 1024 * 1024},
		{PingData: &PingData{IP: &net.IPAddr{IP: net.ParseIP("1.1.1.3")}, Sended: 4, Received: 4, Delay: 12 * time.Millisecond}, DownloadSpeed: 10 * 1024 * 1024},
	}

	ExportCsv(data)

	raw, err := os.ReadFile(Output)
	if err != nil {
		t.Fatalf("read csv failed: %v", err)
	}
	records, err := csv.NewReader(strings.NewReader(string(raw))).ReadAll()
	if err != nil {
		t.Fatalf("parse csv failed: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected 3 csv rows including header, got %d", len(records))
	}
	if records[1][0] != "1.1.1.1" || records[2][0] != "1.1.1.2" {
		t.Fatalf("unexpected exported rows: %v, %v", records[1][0], records[2][0])
	}
}
