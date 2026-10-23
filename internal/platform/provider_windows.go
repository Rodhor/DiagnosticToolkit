package platform

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// IP Helper API constants not exported by x/sys/windows.
const (
	gaaFlagSkipAnycast     = 0x0002
	gaaFlagSkipMulticast   = 0x0004
	gaaFlagIncludeGateways = 0x0080

	ifOperStatusUp = 1
)

type windowsProvider struct{}

func New() Provider { return &windowsProvider{} }

func (p *windowsProvider) Snapshot(ctx context.Context) (Snapshot, error) {
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

	adapters, err := adapterAddresses()
	if err != nil {
		snap.addWarning("adapter configuration unavailable (GetAdaptersAddresses): %v", err)
	} else {
		gw, gwif := defaultGatewayFrom(adapters)
		if gw.IsValid() {
			snap.Gateway = gw
			snap.GatewayIf = gwif
			if snap.GatewayIf == "" {
				snap.GatewayIf = interfaceFor(gw, snap.Interfaces)
			}
		} else {
			snap.addWarning("no default gateway found")
		}

		snap.Resolvers, snap.SearchDomains = resolversFrom(adapters)
		if allLoopback(snap.Resolvers) {
			snap.addWarning("configured resolvers (%v) are a local stub whose upstream servers could not be determined", snap.Resolvers)
		}
	}

	infos, err := adapterInfoByIndex()
	if err != nil {
		snap.addWarning("DHCP lease information unavailable (GetAdaptersInfo): %v", err)
		return snap, nil
	}
	for _, ifi := range snap.Interfaces {
		if ifi.Loopback || !ifi.Up {
			continue
		}
		lease, err := leaseFor(ifi.Name, infos)
		if err != nil {
			snap.addWarning("no DHCP lease for %s: %v", ifi.Name, err)
			continue
		}
		snap.Leases[ifi.Name] = lease
	}

	return snap, nil
}

// adapterAddresses calls GetAdaptersAddresses with the usual
// grow-the-buffer loop and returns the linked list as a slice. The
// entries point into buf, which stays alive as long as they do.
func adapterAddresses() ([]*windows.IpAdapterAddresses, error) {
	const flags = gaaFlagIncludeGateways | gaaFlagSkipAnycast | gaaFlagSkipMulticast
	size := uint32(15000)
	for range 3 {
		buf := make([]byte, size)
		err := windows.GetAdaptersAddresses(
			windows.AF_UNSPEC, flags, 0,
			(*windows.IpAdapterAddresses)(unsafe.Pointer(&buf[0])), &size)
		if err == nil {
			var out []*windows.IpAdapterAddresses
			for aa := (*windows.IpAdapterAddresses)(unsafe.Pointer(&buf[0])); aa != nil; aa = aa.Next {
				out = append(out, aa)
			}
			return out, nil
		}
		if err != windows.ERROR_BUFFER_OVERFLOW {
			return nil, err
		}
	}
	return nil, errors.New("adapter list kept growing between calls")
}

func adapterUsable(aa *windows.IpAdapterAddresses) bool {
	return aa.OperStatus == ifOperStatusUp && aa.IfType != windows.IF_TYPE_SOFTWARE_LOOPBACK
}

// defaultGatewayFrom picks the gateway of the up, non-loopback adapter
// with the lowest interface metric - the same route Windows itself
// prefers when several adapters (wifi + ethernet) have one. IPv4
// gateways win over IPv6 because the checks downstream target IPv4.
func defaultGatewayFrom(adapters []*windows.IpAdapterAddresses) (netip.Addr, string) {
	var (
		best       netip.Addr
		bestIf     string
		bestMetric uint32
		bestV4     bool
	)
	for _, aa := range adapters {
		if !adapterUsable(aa) {
			continue
		}
		name := windows.UTF16PtrToString(aa.FriendlyName)
		for gw := aa.FirstGatewayAddress; gw != nil; gw = gw.Next {
			addr, ok := socketAddrToNetip(&gw.Address)
			if !ok {
				continue
			}
			metric := aa.Ipv6Metric
			if addr.Is4() {
				metric = aa.Ipv4Metric
			}
			better := !best.IsValid() ||
				(addr.Is4() && !bestV4) ||
				(addr.Is4() == bestV4 && metric < bestMetric)
			if better {
				best, bestIf, bestMetric, bestV4 = addr, name, metric, addr.Is4()
			}
		}
	}
	return best, bestIf
}

// resolversFrom collects the DNS servers and connection-specific
// suffixes of every usable adapter, deduplicated in adapter order.
func resolversFrom(adapters []*windows.IpAdapterAddresses) ([]netip.Addr, []string) {
	var (
		resolvers []netip.Addr
		search    []string
	)
	seenAddr := make(map[netip.Addr]bool)
	seenDom := make(map[string]bool)
	for _, aa := range adapters {
		if !adapterUsable(aa) {
			continue
		}
		for dns := aa.FirstDnsServerAddress; dns != nil; dns = dns.Next {
			addr, ok := socketAddrToNetip(&dns.Address)
			if !ok || seenAddr[addr] || isPhantomSiteLocalDNS(addr) {
				continue
			}
			seenAddr[addr] = true
			resolvers = append(resolvers, addr)
		}
		if suffix := windows.UTF16PtrToString(aa.DnsSuffix); suffix != "" && !seenDom[suffix] {
			seenDom[suffix] = true
			search = append(search, suffix)
		}
	}
	return resolvers, search
}

// isPhantomSiteLocalDNS reports whether addr is one of the deprecated
// site-local DNS defaults (fec0:0:0:ffff::1-3) that Windows lists on
// adapters with no real DNS configuration. They answer nothing and
// would pollute the diagnosis.
func isPhantomSiteLocalDNS(addr netip.Addr) bool {
	if !addr.Is6() {
		return false
	}
	return netip.MustParsePrefix("fec0:0:0:ffff::/64").Contains(addr)
}

func socketAddrToNetip(sa *windows.SocketAddress) (netip.Addr, bool) {
	ip := sa.IP()
	if ip == nil {
		return netip.Addr{}, false
	}
	addr, ok := netip.AddrFromSlice(ip)
	if !ok || addr.IsUnspecified() {
		return netip.Addr{}, false
	}
	return addr.Unmap(), true
}

// adapterInfoByIndex calls the legacy GetAdaptersInfo, the only IP
// Helper call that exposes DHCP lease times. IPv4 only, which is all
// DHCP covers anyway.
func adapterInfoByIndex() (map[uint32]*windows.IpAdapterInfo, error) {
	size := uint32(15000)
	for range 3 {
		buf := make([]byte, size)
		err := windows.GetAdaptersInfo(
			(*windows.IpAdapterInfo)(unsafe.Pointer(&buf[0])), &size)
		if err == nil {
			out := make(map[uint32]*windows.IpAdapterInfo)
			for ai := (*windows.IpAdapterInfo)(unsafe.Pointer(&buf[0])); ai != nil; ai = ai.Next {
				out[ai.Index] = ai
			}
			return out, nil
		}
		if err != windows.ERROR_BUFFER_OVERFLOW {
			return nil, err
		}
	}
	return nil, errors.New("adapter list kept growing between calls")
}

func leaseFor(ifname string, infos map[uint32]*windows.IpAdapterInfo) (*Lease, error) {
	ni, err := net.InterfaceByName(ifname)
	if err != nil {
		return nil, err
	}
	info, ok := infos[uint32(ni.Index)]
	if !ok {
		return nil, fmt.Errorf("no adapter info for interface index %d", ni.Index)
	}
	if info.DhcpEnabled == 0 {
		return nil, errors.New("DHCP not enabled")
	}

	lease := &Lease{Source: "GetAdaptersInfo (IP Helper API)"}
	if addr, ok := parseAdapterIPString(info.DhcpServer.IpAddress.String); ok {
		lease.Server = addr
	}
	if addr, ok := parseAdapterIPString(info.GatewayList.IpAddress.String); ok {
		lease.Router = addr
	}
	if info.LeaseObtained != 0 {
		lease.Obtained = time.Unix(info.LeaseObtained, 0)
	}
	if info.LeaseExpires != 0 {
		lease.Expires = time.Unix(info.LeaseExpires, 0)
	}
	return lease, nil
}

// parseAdapterIPString decodes the NUL-terminated dotted-quad strings
// GetAdaptersInfo uses. "0.0.0.0" means unset.
func parseAdapterIPString(raw [16]byte) (netip.Addr, bool) {
	end := 0
	for end < len(raw) && raw[end] != 0 {
		end++
	}
	addr, err := netip.ParseAddr(string(raw[:end]))
	if err != nil || addr.IsUnspecified() {
		return netip.Addr{}, false
	}
	return addr, true
}
