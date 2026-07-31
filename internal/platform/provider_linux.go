package platform

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"os"
	"time"
)

type linuxProvider struct{}

func New() Provider { return &linuxProvider{} }

func (p *linuxProvider) Snapshot(ctx context.Context) (Snapshot, error) {
	snap := Snapshot{
		TakenAt: time.Now(),
		Leases:  make(map[string]*Lease),
	}

	if host, err := os.Hostname(); err == nil {
		snap.Hostname = host
	} else {
		snap.addWarning("hostname unavailable: %v", err)
	}

	// Interfaces are the only hard requirement - without them nothing
	// else in the snapshot is meaningful.
	ifaces, err := readInterfaces()
	if err != nil {
		return Snapshot{}, fmt.Errorf("reading interfaces: %w", err)
	}
	snap.Interfaces = ifaces

	gw, err := readDefaultRoute()
	if err != nil {
		snap.addWarning("default route unavailable: %v", err)
	} else {
		snap.Gateway = gw
		snap.GatewayIf = interfaceFor(gw, snap.Interfaces)
	}

	rc, err := readResolvers()
	if err != nil {
		snap.addWarning("resolver configuration unavailable: %v", err)
	} else {
		snap.Resolvers = rc.Nameservers
		snap.SearchDomains = rc.Search
		if allLoopback(rc.Nameservers) {
			up, err := readUpstreamResolvers()
			if err != nil {
				snap.addWarning("resolv.conf points at a local stub (%v) and the upstream servers could not be determined: %v", rc.Nameservers, err)
			} else {
				snap.UpstreamDNS = up
			}
		}
	}

	for _, ifi := range snap.Interfaces {
		if ifi.Loopback || !ifi.Up {
			continue
		}
		lease, err := readLease(ifi.Name)
		if err != nil {
			snap.addWarning("no DHCP lease for %s: %v", ifi.Name, err)
			continue
		}
		snap.Leases[ifi.Name] = lease
	}

	return snap, nil
}

func allLoopback(addrs []netip.Addr) bool {
	if len(addrs) == 0 {
		return false
	}
	for _, a := range addrs {
		if !a.IsLoopback() {
			return false
		}
	}
	return true
}

// interfaceFor returns the name of the interface whose address prefix
// contains addr, or "" if none does. The kernel route table only gives
// the gateway address; the owning interface has to be derived.
func interfaceFor(addr netip.Addr, ifaces []Interface) string {
	for _, ifi := range ifaces {
		for _, p := range ifi.Addresses {
			if p.Contains(addr) {
				return ifi.Name
			}
		}
	}
	return ""
}

func readInterfaces() ([]Interface, error) {
	raw, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	out := make([]Interface, 0, len(raw))
	for _, ifi := range raw {
		item := Interface{
			Name:     ifi.Name,
			MAC:      ifi.HardwareAddr.String(),
			Up:       ifi.Flags&net.FlagUp != 0,
			Loopback: ifi.Flags&net.FlagLoopback != 0,
			MTU:      ifi.MTU,
		}

		// A failure on one interface should not drop the interface;
		// it is still worth reporting without addresses.
		addrs, err := ifi.Addrs()
		if err != nil {
			out = append(out, item)
			continue
		}

		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			addr, ok := netip.AddrFromSlice(ipnet.IP)
			if !ok {
				continue
			}
			// Unmap: net.IPNet.IP may hold IPv4 in 16-byte v4-mapped
			// form, which would render as ::ffff:192.168.1.42 and
			// compare unequal to the same address parsed from a string.
			addr = addr.Unmap()
			ones, _ := ipnet.Mask.Size()
			// If the mask came in 16-byte form for a v4 address, the
			// bit count includes the 96-bit v4-mapped prefix.
			if addr.Is4() && ones > 32 {
				ones -= 96
			}
			item.Addresses = append(item.Addresses, netip.PrefixFrom(addr, ones))
		}
		out = append(out, item)
	}
	return out, nil
}
