package platform

import (
	"context"
	"errors"
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
		// An interface that is up but has no addresses has no lease to
		// find; a stale lease file from a previous connection is not
		// useful information.
		if ifi.Loopback || !ifi.Up || len(ifi.Addresses) == 0 {
			continue
		}
		lease, err := readLease(ifi.Name)
		if err != nil {
			switch {
			case errors.Is(err, ErrLeaseUnavailable):
				snap.addWarning("DHCP lease detail unavailable for %s (NetworkManager does not expose it on disk)", ifi.Name)
			case errors.Is(err, ErrLeasePermission):
				snap.addWarning("DHCP lease for %s requires root; run with sudo to include it", ifi.Name)
			default:
				snap.addWarning("no DHCP lease for %s: %v", ifi.Name, err)
			}
			continue
		}
		snap.Leases[ifi.Name] = lease
	}

	return snap, nil
}
