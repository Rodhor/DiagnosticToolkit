package app

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/rodhor/DiagnosticToolset/internal/check"
	"github.com/rodhor/DiagnosticToolset/internal/platform"
)

func gatewayTarget(snap platform.Snapshot) string {
	if !snap.Gateway.IsValid() {
		return ""
	}
	return snap.Gateway.String()
}

func discoveredResolvers(configured []check.ResolverSpec, snap platform.Snapshot) []check.ResolverSpec {
	seen := make(map[string]bool, len(configured))
	for _, s := range configured {
		seen[normalizeResolverAddr(s.Address)] = true
	}

	var out []check.ResolverSpec

	add := func(addrs []netip.Addr, label string) {
		for _, a := range addrs {
			addr := net.JoinHostPort(a.String(), "53")
			if seen[addr] {
				continue
			}
			seen[addr] = true
			out = append(out, check.ResolverSpec{
				Name:    fmt.Sprintf("%s(%s)", label, a),
				Address: addr,
			})
		}
	}

	add(snap.Resolvers, "Discovered")
	add(snap.UpstreamDNS, "Upstream")

	return out
}

func normalizeResolverAddr(addr string) string {
	if addr == "" {
		return ""
	}
	if _, _, err := net.SplitHostPort(addr); err != nil {
		return net.JoinHostPort(addr, "53")
	}

	return addr

}
