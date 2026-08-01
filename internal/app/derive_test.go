package app

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

func TestGatewayTarget(t *testing.T) {
	t.Run("empty snapshot yields no target", func(t *testing.T) {
		snap := snapshotFrom(t, &platform.FakeProvider{})
		if got := gatewayTarget(snap); got != "" {
			t.Errorf("gatewayTarget = %q, want empty", got)
		}
	})

	t.Run("gateway returned", func(t *testing.T) {
		snap := snapshotFrom(t, &platform.FakeProvider{Result: platform.Snapshot{
			Gateway: netip.MustParseAddr("192.168.1.1"),
		}})
		if got := gatewayTarget(snap); got != "192.168.1.1" {
			t.Errorf("gatewayTarget = %q, want %q", got, "192.168.1.1")
		}
	})
}

func TestDiscoveredResolvers(t *testing.T) {
	configured := []check.ResolverSpec{
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

	t.Run("empty snapshot yields nothing", func(t *testing.T) {
		snap := snapshotFrom(t, &platform.FakeProvider{})
		if got := discoveredResolvers(configured, snap); len(got) != 0 {
			t.Errorf("specs = %v, want none", names(got))
		}
	})

	t.Run("resolvers and upstream returned with port and label", func(t *testing.T) {
		snap := snapshotFrom(t, &platform.FakeProvider{Result: platform.Snapshot{
			Resolvers:   []netip.Addr{netip.MustParseAddr("127.0.0.53")},
			UpstreamDNS: []netip.Addr{netip.MustParseAddr("192.168.1.1")},
		}})

		got := discoveredResolvers(configured, snap)
		want := []string{"Discovered(127.0.0.53)", "Upstream(192.168.1.1)"}
		if !slices.Equal(names(got), want) {
			t.Fatalf("specs = %v, want %v", names(got), want)
		}
		if got[0].Address != "127.0.0.53:53" {
			t.Errorf("address = %q, want %q", got[0].Address, "127.0.0.53:53")
		}
	})

	t.Run("duplicates of configured resolvers skipped", func(t *testing.T) {
		snap := snapshotFrom(t, &platform.FakeProvider{Result: platform.Snapshot{
			Resolvers: []netip.Addr{netip.MustParseAddr("1.1.1.1")},
		}})
		if got := discoveredResolvers(configured, snap); len(got) != 0 {
			t.Errorf("specs = %v, want none", names(got))
		}
	})
}
