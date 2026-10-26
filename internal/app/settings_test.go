package app

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/rodhor/DiagnosticToolset/internal/check"
)

func TestSettingsViewRoundTrip(t *testing.T) {
	cfg := Default()
	cfg.TCPTargets = []string{"1.1.1.1:53"}
	cfg.PingTargets = []string{"8.8.8.8"}
	cfg.Resolvers = []check.ResolverSpec{{Name: "System"}, {Name: "Cloudflare", Address: "1.1.1.1"}}
	cfg.Queries = []check.QuerySpec{
		{Name: "one.one.one.one", ExpectedOutcome: check.ExpectedSuccess, ExpectedAddresses: []string{"1.1.1.1"}},
		{Name: "not-a-domain.example", ExpectedOutcome: check.ExpectedNXDomain},
	}

	got, err := ViewToConfig(ConfigToView(cfg), Default())
	if err != nil {
		t.Fatalf("ViewToConfig: %v", err)
	}
	if !slices.Equal(got.Resolvers, cfg.Resolvers) || !sameQueries(got.Queries, cfg.Queries) {
		t.Fatalf("DNS settings did not round trip: resolvers=%+v queries=%+v", got.Resolvers, got.Queries)
	}

	path := filepath.Join(t.TempDir(), "netcheck.yaml")
	if err := SaveConfig(path, got); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load saved config: %v", err)
	}
	if !slices.Equal(loaded.Resolvers, cfg.Resolvers) || !sameQueries(loaded.Queries, cfg.Queries) {
		t.Fatalf("saved DNS settings did not round trip: resolvers=%+v queries=%+v", loaded.Resolvers, loaded.Queries)
	}
}

func sameQueries(a, b []check.QuerySpec) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name || a[i].ExpectedOutcome != b[i].ExpectedOutcome || !slices.Equal(a[i].ExpectedAddresses, b[i].ExpectedAddresses) {
			return false
		}
	}
	return true
}

func TestViewToConfigRejectsInvalidSettings(t *testing.T) {
	base := Default()
	valid := ConfigToView(base)
	cases := []struct {
		name   string
		change func(*SettingsView)
	}{
		{"negative timeout", func(v *SettingsView) { v.Timeout = "-1s" }},
		{"negative threshold", func(v *SettingsView) { v.ThresholdDNS = "-1ms" }},
		{"zero ping count", func(v *SettingsView) { v.PingCount = "0" }},
		{"loss thresholds inverted", func(v *SettingsView) { v.PingWarnLoss, v.PingFailLoss = "60", "50" }},
		{"NXDOMAIN with expected addresses", func(v *SettingsView) {
			v.Queries = []QueryView{{Name: "bad.example", Expect: "NXDOMAIN", Expected: "1.1.1.1"}}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			view := valid
			tc.change(&view)
			if _, err := ViewToConfig(view, base); err == nil {
				t.Fatal("ViewToConfig accepted invalid settings")
			}
		})
	}
}
