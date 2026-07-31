package platform

import (
	"net/netip"
	"strings"
	"testing"
	"time"
)

func TestParseFlatLease(t *testing.T) {
	mtime := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		input       string
		wantServer  netip.Addr
		wantRouter  netip.Addr
		wantExpires time.Time
	}{
		{
			name: "networkmanager EXPIRY form",
			input: "# This is private data. Do not parse.\n" +
				"ADDRESS=192.168.1.42\n" +
				"SERVER_ADDRESS=192.168.1.1\n" +
				"ROUTER=192.168.1.1\n" +
				"EXPIRY=1785600000\n",
			wantServer:  netip.MustParseAddr("192.168.1.1"),
			wantRouter:  netip.MustParseAddr("192.168.1.1"),
			wantExpires: time.Unix(1785600000, 0),
		},
		{
			name: "networkd LIFETIME form",
			input: "SERVER_ADDRESS=10.0.0.1\n" +
				"ROUTER=10.0.0.1\n" +
				"LIFETIME=3600\n",
			wantServer:  netip.MustParseAddr("10.0.0.1"),
			wantRouter:  netip.MustParseAddr("10.0.0.1"),
			wantExpires: mtime.Add(3600 * time.Second),
		},
		{
			name: "both present - EXPIRY wins",
			input: "EXPIRY=1785600000\n" +
				"LIFETIME=3600\n",
			wantExpires: time.Unix(1785600000, 0),
		},
		{
			name:  "empty file",
			input: "",
		},
		{
			name: "malformed lines ignored",
			input: "this is not a key value pair\n" +
				"=value-without-key\n" +
				"SERVER_ADDRESS=not-an-ip\n" +
				"router=172.16.0.1\n", // lowercase key still accepted
			wantRouter: netip.MustParseAddr("172.16.0.1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lease, err := parseFlatLease(strings.NewReader(tt.input), mtime)
			if err != nil {
				t.Fatalf("parseFlatLease: %v", err)
			}
			if lease.Server != tt.wantServer {
				t.Errorf("Server = %v, want %v", lease.Server, tt.wantServer)
			}
			if lease.Router != tt.wantRouter {
				t.Errorf("Router = %v, want %v", lease.Router, tt.wantRouter)
			}
			if !lease.Expires.Equal(tt.wantExpires) {
				t.Errorf("Expires = %v, want %v", lease.Expires, tt.wantExpires)
			}
			if !lease.Obtained.Equal(mtime) {
				t.Errorf("Obtained = %v, want mtime %v", lease.Obtained, mtime)
			}
		})
	}
}
