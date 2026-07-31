package platform

import (
	"bufio"
	"io"
	"net/netip"
	"os"
	"strings"
)

const (
	resolvConfPath = "/etc/resolv.conf"
	// systemd-resolved writes the real upstream servers here when
	// /etc/resolv.conf only lists its 127.0.0.53 stub listener.
	upstreamResolvConfPath = "/run/systemd/resolve/resolv.conf"
)

type resolvConf struct {
	Nameservers []netip.Addr
	Search      []string
}

// parseResolvConf handles "nameserver <ip>", "search <domains...>" and
// the legacy single "domain <name>" form. Comments start with # or ;.
func parseResolvConf(r io.Reader) (resolvConf, error) {
	var rc resolvConf

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if i := strings.IndexAny(line, "#;"); i >= 0 {
			line = line[:i]
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "nameserver":
			raw := fields[1]
			// Strip zone suffix (fe80::1%eth0) - netip keeps the zone
			// and it is meaningless for identifying the server.
			if i := strings.IndexByte(raw, '%'); i >= 0 {
				raw = raw[:i]
			}
			addr, err := netip.ParseAddr(raw)
			if err != nil {
				continue
			}
			rc.Nameservers = append(rc.Nameservers, addr)
		case "search":
			rc.Search = append(rc.Search, fields[1:]...)
		case "domain":
			rc.Search = append(rc.Search, fields[1])
		}
	}
	return rc, scanner.Err()
}

func readResolvers() (resolvConf, error) {
	f, err := os.Open(resolvConfPath)
	if err != nil {
		return resolvConf{}, err
	}
	defer f.Close()
	return parseResolvConf(f)
}

func readUpstreamResolvers() ([]netip.Addr, error) {
	f, err := os.Open(upstreamResolvConfPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rc, err := parseResolvConf(f)
	if err != nil {
		return nil, err
	}
	return rc.Nameservers, nil
}
