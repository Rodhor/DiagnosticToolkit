package platform

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ErrLeasePermission indicates a lease file or its directory exists but
// could not be read with the current privileges.
var ErrLeasePermission = errors.New("lease requires elevated privileges")

// ErrLeaseUnavailable indicates the DHCP client in use does not expose
// lease detail on disk. NetworkManager records only the assigned address
// and keeps the rest of the lease in internal state, so no amount of
// privilege makes it readable this way.
var ErrLeaseUnavailable = errors.New("DHCP lease detail not exposed by this DHCP client")

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
// turns it into a warning. A permission problem on the networkd path
// is reported as ErrLeasePermission (root genuinely helps there);
// anything NetworkManager-shaped that yields no detail is reported as
// ErrLeaseUnavailable (root would not help).
func readLease(ifname string) (*Lease, error) {
	lease, err := readNetworkdLease(ifname)
	if err == nil {
		return lease, nil
	}
	if errors.Is(err, fs.ErrPermission) {
		return nil, fmt.Errorf("systemd-networkd: %w", ErrLeasePermission)
	}

	lease, err = readNetworkManagerLease(ifname)
	if err == nil {
		return lease, nil
	}
	// A permission error here is NOT a reason to suggest sudo: even with
	// access, NetworkManager's on-disk file holds only the assigned
	// address (see readNetworkManagerLease), so privilege gains nothing.
	if errors.Is(err, fs.ErrPermission) || errors.Is(err, ErrLeaseUnavailable) {
		return nil, fmt.Errorf("NetworkManager: %w", ErrLeaseUnavailable)
	}

	return nil, errors.New("no lease file found (checked systemd-networkd and NetworkManager)")
}

func readNetworkdLease(ifname string) (*Lease, error) {
	ifi, err := net.InterfaceByName(ifname)
	if err != nil {
		return nil, err
	}
	return readLeaseFile(filepath.Join(networkdLeaseDir, strconv.Itoa(ifi.Index)))
}

func readNetworkManagerLease(ifname string) (*Lease, error) {
	// filepath.Glob reports an unreadable directory as zero matches,
	// not an error, so probe readability first to tell "no lease"
	// apart from "not allowed to look".
	dir, err := os.Open(networkManagerLeaseDir)
	if err != nil {
		return nil, err
	}
	dir.Close()

	pattern := filepath.Join(networkManagerLeaseDir, "internal-*-"+ifname+".lease")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	// One lease file exists per connection profile the interface has
	// ever joined; the most recently modified one is the active
	// connection.
	path := newestFile(matches)
	if path == "" {
		return nil, fmt.Errorf("no lease file matching %s: %w", pattern, fs.ErrNotExist)
	}
	lease, err := readLeaseFile(path)
	if err != nil {
		return nil, err
	}
	// NetworkManager's file typically contains nothing but
	// "ADDRESS=x.x.x.x" - the real lease lives in NetworkManager's
	// internal state, reachable only over D-Bus. A struct holding just
	// the file's mtime and zero values is not a lease.
	if !hasLeaseDetail(lease) {
		return nil, fmt.Errorf("%s: %w", path, ErrLeaseUnavailable)
	}
	return lease, nil
}

// hasLeaseDetail reports whether the lease carries at least one
// meaningful field beyond the assigned address.
func hasLeaseDetail(lease *Lease) bool {
	return lease.Server.IsValid() || lease.Router.IsValid() || !lease.Expires.IsZero()
}

// newestFile returns the path with the most recent modification time,
// skipping entries that cannot be stat'd. Returns "" if none remain.
func newestFile(paths []string) string {
	var (
		newest string
		mtime  time.Time
	)
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if newest == "" || info.ModTime().After(mtime) {
			newest, mtime = path, info.ModTime()
		}
	}
	return newest
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
