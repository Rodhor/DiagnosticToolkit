package platform

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"os"
	"strconv"
	"strings"
)

const procNetRoute = "/proc/net/route"

// Route flag bits from the kernel (include/uapi/linux/route.h).
const (
	routeFlagUp      = 0x0001
	routeFlagGateway = 0x0002
)

var errNoDefaultRoute = errors.New("no default route found")

// parseRoutes reads /proc/net/route content and returns the gateway of
// the default route with the lowest metric. A machine on both wifi and
// ethernet has multiple default routes; taking the first row would give
// a plausible but wrong answer.
func parseRoutes(r io.Reader) (netip.Addr, error) {
	var (
		best       netip.Addr
		bestMetric int
		found      bool
	)

	scanner := bufio.NewScanner(r)
	first := true
	for scanner.Scan() {
		if first { // header row
			first = false
			continue
		}
		fields := strings.Fields(scanner.Text())
		// Iface Destination Gateway Flags RefCnt Use Metric Mask ...
		if len(fields) < 7 {
			continue
		}
		if fields[1] != "00000000" {
			continue
		}
		flags, err := strconv.ParseUint(fields[3], 16, 32)
		if err != nil || flags&(routeFlagUp|routeFlagGateway) != routeFlagUp|routeFlagGateway {
			continue
		}
		gw, err := parseLittleEndianIPv4(fields[2])
		if err != nil {
			continue
		}
		metric, err := strconv.Atoi(fields[6])
		if err != nil {
			continue
		}
		if !found || metric < bestMetric {
			best, bestMetric, found = gw, metric, true
		}
	}
	if err := scanner.Err(); err != nil {
		return netip.Addr{}, err
	}
	if !found {
		return netip.Addr{}, errNoDefaultRoute
	}
	return best, nil
}

// parseLittleEndianIPv4 decodes the hex address format of
// /proc/net/route: "0101A8C0" -> 192.168.1.1.
func parseLittleEndianIPv4(s string) (netip.Addr, error) {
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("bad address %q: %w", s, err)
	}
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], uint32(v))
	return netip.AddrFrom4(b), nil
}

func readDefaultRoute() (netip.Addr, error) {
	f, err := os.Open(procNetRoute)
	if err != nil {
		return netip.Addr{}, err
	}
	defer f.Close()
	return parseRoutes(f)
}
