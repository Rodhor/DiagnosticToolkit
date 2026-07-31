package platform

import (
	"context"
	"fmt"
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
