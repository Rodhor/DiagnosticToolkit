package pages

import (
	"fmt"
	"net/netip"
	"sort"
	"strings"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/app"
	"github.com/rodhor/DiagnosticToolset/internal/domain"
	"github.com/rodhor/DiagnosticToolset/internal/platform"
	"github.com/rodhor/DiagnosticToolset/internal/web/server"
)

func fmtInt(n int) string {
	return fmt.Sprintf("%d", n)
}

func fmtFloat(f float64) string {
	return fmt.Sprintf("%g", f)
}

func qName(qs []app.QueryView, i int) string {
	if i < len(qs) {
		return qs[i].Name
	}
	return ""
}

func qExpected(qs []app.QueryView, i int) string {
	if i < len(qs) {
		return qs[i].Expected
	}
	return ""
}

func qExpect(qs []app.QueryView, i int) string {
	if i < len(qs) {
		return qs[i].Expect
	}
	return ""
}

func qRowCount(qs []app.QueryView) int {
	n := len(qs) + 1
	if n < 3 {
		n = 3
	}
	return n
}

func runHasFailure(r *server.Run) bool {
	if r == nil || r.Buffer == nil {
		return false
	}
	for _, result := range r.Buffer.All() {
		if result.Status == domain.Fail {
			return true
		}
	}
	return false
}

func runHasWarning(r *server.Run) bool {
	if r == nil || r.Buffer == nil {
		return false
	}
	for _, result := range r.Buffer.All() {
		if result.Status == domain.Warn {
			return true
		}
	}
	return false
}

func runHasUnknown(r *server.Run) bool {
	if r == nil || r.Buffer == nil {
		return false
	}
	for _, result := range r.Buffer.All() {
		if result.Status == domain.Unknown {
			return true
		}
	}
	return false
}

type interfaceView struct {
	Name      string
	State     string
	Addresses string
	MAC       string
	MTU       int
	Lease     string
}

type environmentView struct {
	Hostname      string
	Captured      string
	Gateway       string
	GatewayIf     string
	Resolvers     []string
	UpstreamDNS   []string
	SearchDomains []string
	Interfaces    []interfaceView
	Warnings      []string
}

func toEnvironmentView(s platform.Snapshot) environmentView {
	v := environmentView{
		Hostname:      s.Hostname,
		Captured:      s.TakenAt.Format(time.RFC1123),
		Gateway:       addrOrUnknown(s.Gateway),
		GatewayIf:     s.GatewayIf,
		Resolvers:     addrStrings(s.Resolvers),
		UpstreamDNS:   addrStrings(s.UpstreamDNS),
		SearchDomains: append([]string(nil), s.SearchDomains...),
		Warnings:      append([]string(nil), s.Warnings...),
	}
	v.Interfaces = make([]interfaceView, 0, len(s.Interfaces))
	for _, nic := range s.Interfaces {
		state := "Down"
		if nic.Up {
			state = "Up"
		}
		if nic.Loopback {
			state = "Loopback"
		}
		var addresses []string
		for _, address := range nic.Addresses {
			addresses = append(addresses, address.String())
		}
		lease := ""
		if record := s.Leases[nic.Name]; record != nil {
			var parts []string
			if record.Server.IsValid() {
				parts = append(parts, "DHCP server "+record.Server.String())
			}
			if record.Router.IsValid() {
				parts = append(parts, "router "+record.Router.String())
			}
			if !record.Expires.IsZero() {
				parts = append(parts, "expires "+record.Expires.Format("Jan 02 15:04"))
			}
			lease = strings.Join(parts, " · ")
		}
		v.Interfaces = append(v.Interfaces, interfaceView{
			Name: nic.Name, State: state, Addresses: strings.Join(addresses, ", "),
			MAC: nic.MAC, MTU: nic.MTU, Lease: lease,
		})
	}
	sort.Slice(v.Interfaces, func(i, j int) bool { return v.Interfaces[i].Name < v.Interfaces[j].Name })
	sort.Strings(v.Resolvers)
	sort.Strings(v.UpstreamDNS)
	sort.Strings(v.SearchDomains)
	return v
}

func addrStrings(addrs []netip.Addr) []string {
	out := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		if addr.IsValid() {
			out = append(out, addr.String())
		}
	}
	return out
}

func addrOrUnknown(addr netip.Addr) string {
	if !addr.IsValid() {
		return "Not detected"
	}
	return addr.String()
}
