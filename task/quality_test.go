package task

import (
	"testing"

	"github.com/XIU2/CloudflareSpeedTest/utils"
)

func TestShouldAbortByLossRate(t *testing.T) {
	originalLossRate := utils.InputMaxLossRate
	originalPingTimes := PingTimes
	t.Cleanup(func() {
		utils.InputMaxLossRate = originalLossRate
		PingTimes = originalPingTimes
	})

	PingTimes = 4
	utils.InputMaxLossRate = 0.25
	if shouldAbortByLossRate(1) {
		t.Fatalf("expected one failure at 25%% loss to still be allowed")
	}
	if !shouldAbortByLossRate(2) {
		t.Fatalf("expected two failures at 25%% loss to abort")
	}

	utils.InputMaxLossRate = 1.0
	if shouldAbortByLossRate(10) {
		t.Fatalf("expected default max loss rate to disable early abort")
	}
}
