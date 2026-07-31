package main

import (
	"slices"
	"fmt"
	"net"
	"net/netip"

	"github.com/rodhor/DiagnosticToolset/internal/check"
	"github.com/rodhor/DiagnosticToolset/internal/platform"
)

// appendGatewayTarget adds the discovered gateway to the ping targets.
// Discovery adds to the hardcoded targets; it never replaces them.
func appendGatewayTarget(targets []string, snap platform.Snapshot) []string {
	if !snap.Gateway.IsValid() {
		return targets
	}
	gw := snap.Gateway.String()
	if slices.Contains(targets, gw) {
			return targets
		}
	return append(targets, gw)
}

// appendDiscoveredResolvers adds the snapshot's resolvers and upstream
// DNS servers to the DNS check's resolver list. The hardcoded public
// resolvers stay - the diagnostic value of the DNS matrix comes from
// comparing the customer's resolver against a known-good one.
func appendDiscoveredResolvers(specs []check.ResolverSpec, snap platform.Snapshot) []check.ResolverSpec {
	seen := make(map[string]bool, len(specs))
	for _, s := range specs {
		seen[normalizeResolverAddr(s.Address)] = true
	}

	add := func(addrs []netip.Addr, label string) {
		for _, a := range addrs {
			addr := net.JoinHostPort(a.String(), "53")
			if seen[addr] {
				continue
			}
			seen[addr] = true
			specs = append(specs, check.ResolverSpec{
				Name:    fmt.Sprintf("%s(%s)", label, a),
				Address: addr,
			})
		}
	}
	add(snap.Resolvers, "Discovered")
	add(snap.UpstreamDNS, "Upstream")
	return specs
}

// normalizeResolverAddr mirrors NewDNSCheck's defaulting of the port
// so duplicate detection compares like with like.
func normalizeResolverAddr(addr string) string {
	if addr == "" {
		return ""
	}
	if _, _, err := net.SplitHostPort(addr); err != nil {
		return net.JoinHostPort(addr, "53")
	}
	return addr
}
