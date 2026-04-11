package task

import (
	"net"
	"testing"
)

func TestRandomIPv6StaysInCIDR(t *testing.T) {
	_, ipNet, err := net.ParseCIDR("2606:4700::/32")
	if err != nil {
		t.Fatalf("ParseCIDR failed: %v", err)
	}

	targetIP := randomIPv6(ipNet.IP, ipNet.Mask)
	if !ipNet.Contains(targetIP) {
		t.Fatalf("expected %s to stay in %s", targetIP.String(), ipNet.String())
	}
	if targetIP.To4() != nil {
		t.Fatalf("expected IPv6 address, got %s", targetIP.String())
	}
}

func TestChooseIPv6AddsSingleAddress(t *testing.T) {
	ranges := newIPRanges()
	ranges.parseCIDR("2606:4700::/32")
	ranges.chooseIPv6()

	if len(ranges.ips) != 1 {
		t.Fatalf("expected one IPv6 address, got %d", len(ranges.ips))
	}
	if !ranges.ipNet.Contains(ranges.ips[0].IP) {
		t.Fatalf("expected %s to stay in %s", ranges.ips[0].IP.String(), ranges.ipNet.String())
	}
}