package platform

import (
	"errors"
	"net/netip"
	"strings"
	"testing"
)

const routeHeader = "Iface\tDestination\tGateway \tFlags\tRefCnt\tUse\tMetric\tMask\t\tMTU\tWindow\tIRTT\n"

func routeRow(iface, dest, gw, flags, metric string) string {
	return iface + "\t" + dest + "\t" + gw + "\t" + flags + "\t0\t0\t" + metric + "\t00000000\t0\t0\t0\n"
}

func TestParseRoutes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    netip.Addr
		wantErr error
	}{
		{
			name: "little-endian decoding",
			// 0101A8C0 -> 192.168.1.1
			input: routeHeader + routeRow("eth0", "00000000", "0101A8C0", "0003", "100"),
			want:  netip.MustParseAddr("192.168.1.1"),
		},
		{
			name: "multiple default routes - lowest metric wins",
			input: routeHeader +
				routeRow("wlan0", "00000000", "0100000A", "0003", "600") + // 10.0.0.1
				routeRow("eth0", "00000000", "0101A8C0", "0003", "100"), // 192.168.1.1
			want: netip.MustParseAddr("192.168.1.1"),
		},
		{
			name: "non-default routes ignored",
			input: routeHeader +
				routeRow("eth0", "0001A8C0", "00000000", "0001", "0") +
				routeRow("eth0", "00000000", "0101A8C0", "0003", "100"),
			want: netip.MustParseAddr("192.168.1.1"),
		},
		{
			name: "routes without gateway flag ignored",
			input: routeHeader +
				routeRow("eth0", "00000000", "0100000A", "0001", "0") + // up but not RTF_GATEWAY
				routeRow("eth0", "00000000", "0101A8C0", "0003", "100"),
			want: netip.MustParseAddr("192.168.1.1"),
		},
		{
			name: "malformed rows skipped",
			input: routeHeader +
				"garbage row\n" +
				routeRow("eth0", "00000000", "ZZZZZZZZ", "0003", "100") +
				routeRow("eth0", "00000000", "0101A8C0", "notahex", "100") +
				routeRow("eth0", "00000000", "0101A8C0", "0003", "100"),
			want: netip.MustParseAddr("192.168.1.1"),
		},
		{
			name:    "no default route",
			input:   routeHeader + routeRow("eth0", "0001A8C0", "00000000", "0001", "0"),
			wantErr: errNoDefaultRoute,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: errNoDefaultRoute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRoutes(strings.NewReader(tt.input))
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseRoutes: %v", err)
			}
			if got != tt.want {
				t.Errorf("gateway = %v, want %v", got, tt.want)
			}
		})
	}
}
