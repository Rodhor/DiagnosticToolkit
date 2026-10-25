package platform

import (
	"fmt"
	"net/netip"
	"time"
)

// Interface describes one network adapter as the OS reports it
type Interface struct {
	Name      string
	MAC       string
	Up        bool
	Loopback  bool
	MTU       int
	Addresses []netip.Prefix
}

// Lease describes a DHCP lease. Every field is optional - platforms
// differ wildly in what they expose, and absent is not zero
type Lease struct {
	Server   netip.Addr
	Router   netip.Addr
	Obtained time.Time
	Expires  time.Time
	Source   string // path or mechanism the lease came from
}

// Snapshot is the complete network environment at a point in time
type Snapshot struct {
	Hostname      string
	Interfaces    []Interface
	Gateway       netip.Addr
	GatewayIf     string
	Resolvers     []netip.Addr
	UpstreamDNS   []netip.Addr
	SearchDomains []string
	Leases        map[string]*Lease
	Warnings      []string
	TakenAt       time.Time
}

func (s *Snapshot) addWarning(format string, args ...any) {
	s.Warnings = append(s.Warnings, fmt.Sprintf(format, args...))
}
