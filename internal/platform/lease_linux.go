package platform

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	networkdLeaseDir       = "/run/systemd/netif/leases"
	networkManagerLeaseDir = "/var/lib/NetworkManager"
)

// parseFlatLease reads the flat KEY=value lease format shared by
// systemd-networkd and NetworkManager. Unknown keys and malformed
// lines are ignored - the two writers emit different key sets and
// both may grow new ones.
//
// mtime is the lease file's modification time: it doubles as the
// obtained time and as the base for LIFETIME (seconds, networkd).
// EXPIRY (absolute unix timestamp, NetworkManager) wins when both
// are present.
func parseFlatLease(r io.Reader, mtime time.Time) (*Lease, error) {
	kv := make(map[string]string)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		kv[strings.ToUpper(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	lease := &Lease{Obtained: mtime}

	if v, ok := kv["SERVER_ADDRESS"]; ok {
		if addr, err := netip.ParseAddr(v); err == nil {
			lease.Server = addr
		}
	}
	if v, ok := kv["ROUTER"]; ok {
		if addr, err := netip.ParseAddr(v); err == nil {
			lease.Router = addr
		}
	}

	if v, ok := kv["EXPIRY"]; ok {
		if secs, err := strconv.ParseInt(v, 10, 64); err == nil {
			lease.Expires = time.Unix(secs, 0)
		}
	}
	if !lease.Expires.IsZero() {
		return lease, nil
	}
	if v, ok := kv["LIFETIME"]; ok {
		if secs, err := strconv.ParseInt(v, 10, 64); err == nil {
			lease.Expires = mtime.Add(time.Duration(secs) * time.Second)
		}
	}
	return lease, nil
}

// readLease finds a DHCP lease file for the interface, trying
// systemd-networkd first, then NetworkManager. Failure is expected
// and normal (root-owned paths, other DHCP clients); the caller
// turns it into a warning.
func readLease(ifname string) (*Lease, error) {
	var paths []string

	if ifi, err := net.InterfaceByName(ifname); err == nil {
		paths = append(paths, filepath.Join(networkdLeaseDir, strconv.Itoa(ifi.Index)))
	}

	pattern := filepath.Join(networkManagerLeaseDir, "internal-*-"+ifname+".lease")
	if matches, err := filepath.Glob(pattern); err == nil {
		paths = append(paths, matches...)
	}

	var lastErr error
	for _, path := range paths {
		lease, err := readLeaseFile(path)
		if err != nil {
			lastErr = err
			continue
		}
		return lease, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("no lease file found for %s", ifname)
}

func readLeaseFile(path string) (*Lease, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	lease, err := parseFlatLease(f, info.ModTime())
	if err != nil {
		return nil, err
	}
	lease.Source = path
	return lease, nil
}
