package task

import "github.com/XIU2/CloudflareSpeedTest/utils"

func shouldAbortByLossRate(failedAttempts int) bool {
	if utils.InputMaxLossRate >= 1.0 || PingTimes <= 0 {
		return false
	}
	return float32(failedAttempts)/float32(PingTimes) > utils.InputMaxLossRate
}
