package report

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/rodhor/DiagnosticToolset/internal/platform"
)

// WriteEnvironment renders the discovered network environment before
// the check results. Warnings are always printed - a silently
// incomplete snapshot is worse than one that says what it could not
// determine.
func (cr *ConsoleReport) WriteEnvironment(snap platform.Snapshot) error {
	var b strings.Builder

	fmt.Fprintf(&b, "=== Environment (%s) ===\n", snap.TakenAt.Format("2006-01-02 15:04:05"))

	if snap.Hostname != "" {
		fmt.Fprintf(&b, "Hostname: %s\n", snap.Hostname)
	}

	for _, ifi := range snap.Interfaces {
		if ifi.Loopback || !ifi.Up {
			continue
		}
		fmt.Fprintf(&b, "Interface %s: %s (MAC %s, MTU %d)\n",
			ifi.Name, formatPrefixes(ifi.Addresses), ifi.MAC, ifi.MTU)
	}

	if snap.Gateway.IsValid() {
		fmt.Fprintf(&b, "Gateway: %s", snap.Gateway)
		if snap.GatewayIf != "" {
			fmt.Fprintf(&b, " via %s", snap.GatewayIf)
		}
		b.WriteString("\n")
	} else {
		b.WriteString("Gateway: none found\n")
	}

	if len(snap.Resolvers) > 0 {
		fmt.Fprintf(&b, "Resolvers: %s", formatAddrs(snap.Resolvers))
		if len(snap.UpstreamDNS) > 0 {
			b.WriteString(" (local stub)")
		}
		b.WriteString("\n")
	} else {
		b.WriteString("Resolvers: none found\n")
	}
	if len(snap.UpstreamDNS) > 0 {
		fmt.Fprintf(&b, "Upstream DNS: %s\n", formatAddrs(snap.UpstreamDNS))
	}
	if len(snap.SearchDomains) > 0 {
		fmt.Fprintf(&b, "Search domains: %s\n", strings.Join(snap.SearchDomains, ", "))
	}

	for name, lease := range snap.Leases {
		if lease == nil {
			continue
		}
		fmt.Fprintf(&b, "DHCP lease %s: server %s, router %s, obtained %s, expires %s (%s)\n",
			name,
			formatAddr(lease.Server),
			formatAddr(lease.Router),
			lease.Obtained.Format("2006-01-02 15:04:05"),
			lease.Expires.Format("2006-01-02 15:04:05"),
			lease.Source)
	}

	for _, w := range snap.Warnings {
		fmt.Fprintf(&b, "[WARN] %s\n", w)
	}

	_, err := fmt.Fprintln(cr.writer, b.String())
	return err
}

func formatAddr(a netip.Addr) string {
	if !a.IsValid() {
		return "unknown"
	}
	return a.String()
}

func formatAddrs(addrs []netip.Addr) string {
	parts := make([]string, len(addrs))
	for i, a := range addrs {
		parts[i] = a.String()
	}
	return strings.Join(parts, ", ")
}

func formatPrefixes(prefixes []netip.Prefix) string {
	if len(prefixes) == 0 {
		return "no addresses"
	}
	parts := make([]string, len(prefixes))
	for i, p := range prefixes {
		parts[i] = p.String()
	}
	return strings.Join(parts, ", ")
}
