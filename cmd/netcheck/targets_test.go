package main

import (
	"context"
	"net/netip"
	"slices"
	"testing"

	"github.com/rodhor/DiagnosticToolset/internal/check"
	"github.com/rodhor/DiagnosticToolset/internal/platform"
)

func snapshotFrom(t *testing.T, fake *platform.FakeProvider) platform.Snapshot {
	t.Helper()
	snap, err := fake.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("FakeProvider.Snapshot: %v", err)
	}
	return snap
}

func TestAppendGatewayTarget(t *testing.T) {
	base := []string{"1.1.1.1", "8.8.8.8"}

	t.Run("empty snapshot leaves targets unchanged", func(t *testing.T) {
		snap := snapshotFrom(t, &platform.FakeProvider{})
		got := appendGatewayTarget(base, snap)
		if !slices.Equal(got, base) {
			t.Errorf("targets = %v, want %v", got, base)
		}
	})

	t.Run("gateway appended", func(t *testing.T) {
		snap := snapshotFrom(t, &platform.FakeProvider{Result: platform.Snapshot{
			Gateway: netip.MustParseAddr("192.168.1.1"),
		}})
		got := appendGatewayTarget(base, snap)
		want := []string{"1.1.1.1", "8.8.8.8", "192.168.1.1"}
		if !slices.Equal(got, want) {
			t.Errorf("targets = %v, want %v", got, want)
		}
	})

	t.Run("duplicate gateway not appended", func(t *testing.T) {
		snap := snapshotFrom(t, &platform.FakeProvider{Result: platform.Snapshot{
			Gateway: netip.MustParseAddr("1.1.1.1"),
		}})
		got := appendGatewayTarget(base, snap)
		if !slices.Equal(got, base) {
			t.Errorf("targets = %v, want %v", got, base)
		}
	})
}

func TestAppendDiscoveredResolvers(t *testing.T) {
	base := []check.ResolverSpec{
		{Name: "System", Address: ""},
		{Name: "Cloudflare", Address: "1.1.1.1:53"},
	}

	names := func(specs []check.ResolverSpec) []string {
		out := make([]string, len(specs))
		for i, s := range specs {
			out[i] = s.Name
		}
		return out
	}

	t.Run("empty snapshot leaves specs unchanged", func(t *testing.T) {
		snap := snapshotFrom(t, &platform.FakeProvider{})
		got := appendDiscoveredResolvers(base, snap)
		if len(got) != len(base) {
			t.Errorf("specs = %v, want unchanged %v", names(got), names(base))
		}
	})

	t.Run("resolvers and upstream appended with port and name", func(t *testing.T) {
		snap := snapshotFrom(t, &platform.FakeProvider{Result: platform.Snapshot{
			Resolvers:   []netip.Addr{netip.MustParseAddr("127.0.0.53")},
			UpstreamDNS: []netip.Addr{netip.MustParseAddr("192.168.1.1")},
		}})
		got := appendDiscoveredResolvers(base, snap)
		want := append(names(base),
			"Discovered(127.0.0.53)", "Upstream(192.168.1.1)")
		if !slices.Equal(names(got), want) {
			t.Fatalf("specs = %v, want %v", names(got), want)
		}
		if got[2].Address != "127.0.0.53:53" {
			t.Errorf("discovered address = %q, want %q", got[2].Address, "127.0.0.53:53")
		}
	})

	t.Run("duplicates of hardcoded resolvers skipped", func(t *testing.T) {
		snap := snapshotFrom(t, &platform.FakeProvider{Result: platform.Snapshot{
			Resolvers: []netip.Addr{netip.MustParseAddr("1.1.1.1")},
		}})
		got := appendDiscoveredResolvers(base, snap)
		if len(got) != len(base) {
			t.Errorf("specs = %v, want unchanged %v", names(got), names(base))
		}
	})
}
