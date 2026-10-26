package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/rodhor/DiagnosticToolset/internal/check"
	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

// SettingsView is the form representation of a Config for the web UI.
type QueryView struct {
	Name     string
	Expected string
	Expect   string
}

type SettingsView struct {
	Timeout       string
	Concurrency   string
	Interval      string
	ThresholdTCP  string
	ThresholdDNS  string
	ThresholdPing string

	PingCount    string
	PingInterval string
	PingTimeout  string
	PingWarnLoss string
	PingFailLoss string

	TCPTargets  []string
	PingTargets []string
	Resolvers   []string
	Queries     []QueryView

	OutputResults string
	OutputSummary string
}

func ConfigToView(cfg Config) SettingsView {
	v := SettingsView{
		Timeout:       cfg.Timeout.String(),
		Concurrency:   strconv.Itoa(cfg.Concurrency),
		Interval:      cfg.Interval.String(),
		PingCount:     strconv.Itoa(cfg.Ping.Count),
		PingInterval:  cfg.Ping.Interval.String(),
		PingTimeout:   cfg.Ping.Timeout.String(),
		PingWarnLoss:  strconv.FormatFloat(cfg.Ping.WarnLoss, 'f', -1, 64),
		PingFailLoss:  strconv.FormatFloat(cfg.Ping.FailLoss, 'f', -1, 64),
		TCPTargets:    cfg.TCPTargets,
		PingTargets:   cfg.PingTargets,
		OutputResults: cfg.OutputPath,
		OutputSummary: cfg.SummaryPath,
	}
	if d, ok := cfg.Thresholds[domain.TCP]; ok {
		v.ThresholdTCP = d.String()
	}
	if d, ok := cfg.Thresholds[domain.DNS]; ok {
		v.ThresholdDNS = d.String()
	}
	if d, ok := cfg.Thresholds[domain.Ping]; ok {
		v.ThresholdPing = d.String()
	}
	// Resolvers as tags "name=address"
	if len(cfg.Resolvers) > 0 {
		tags := make([]string, 0, len(cfg.Resolvers))
		for _, r := range cfg.Resolvers {
			if r.Address == "" {
				tags = append(tags, r.Name)
			} else {
				tags = append(tags, r.Name+"="+r.Address)
			}
		}
		v.Resolvers = tags
	}
	// Queries
	if len(cfg.Queries) > 0 {
		qv := make([]QueryView, 0, len(cfg.Queries))
		for _, q := range cfg.Queries {
			exp := ""
			if q.ExpectedOutcome == check.ExpectedNXDomain {
				exp = "NXDOMAIN"
			}
			qv = append(qv, QueryView{
				Name:     q.Name,
				Expected: strings.Join(q.ExpectedAddresses, ", "),
				Expect:   exp,
			})
		}
		v.Queries = qv
	}
	return v
}

func ViewToConfig(v SettingsView, base Config) (Config, error) {
	cfg := base

	d, err := parseDuration(v.Timeout, base.Timeout)
	if err != nil {
		return Config{}, fmt.Errorf("timeout: %w", err)
	}
	cfg.Timeout = d

	concurrency, err := strconv.Atoi(strings.TrimSpace(v.Concurrency))
	if err != nil {
		return Config{}, fmt.Errorf("concurrency must be a whole number")
	}
	if concurrency < 0 {
		return Config{}, fmt.Errorf("concurrency must be non-negative")
	}
	cfg.Concurrency = concurrency

	d, err = parseDuration(v.Interval, base.Interval)
	if err != nil {
		return Config{}, fmt.Errorf("interval: %w", err)
	}
	cfg.Interval = d

	// thresholds
	th := domain.Thresholds{}
	if v.ThresholdTCP != "" {
		dd, err := parseDuration(v.ThresholdTCP, 0)
		if err != nil {
			return Config{}, fmt.Errorf("threshold tcp: %w", err)
		}
		th[domain.TCP] = dd
	}
	if v.ThresholdDNS != "" {
		dd, err := parseDuration(v.ThresholdDNS, 0)
		if err != nil {
			return Config{}, fmt.Errorf("threshold dns: %w", err)
		}
		th[domain.DNS] = dd
	}
	if v.ThresholdPing != "" {
		dd, err := parseDuration(v.ThresholdPing, 0)
		if err != nil {
			return Config{}, fmt.Errorf("threshold ping: %w", err)
		}
		th[domain.Ping] = dd
	}
	if len(th) > 0 {
		cfg.Thresholds = th
	}

	// ping
	cfg.Ping.Count, err = strconv.Atoi(strings.TrimSpace(v.PingCount))
	if err != nil {
		return Config{}, fmt.Errorf("ping count must be a whole number")
	}
	if cfg.Ping.Count <= 0 {
		return Config{}, fmt.Errorf("ping count must be greater than zero")
	}
	if cfg.Ping.Interval, err = parseDuration(v.PingInterval, base.Ping.Interval); err != nil {
		return Config{}, fmt.Errorf("ping interval: %w", err)
	}
	if cfg.Ping.Timeout, err = parseDuration(v.PingTimeout, base.Ping.Timeout); err != nil {
		return Config{}, fmt.Errorf("ping timeout: %w", err)
	}
	cfg.Ping.WarnLoss, err = strconv.ParseFloat(strings.TrimSpace(v.PingWarnLoss), 64)
	if err != nil {
		return Config{}, fmt.Errorf("ping warning loss must be a number")
	}
	cfg.Ping.FailLoss, err = strconv.ParseFloat(strings.TrimSpace(v.PingFailLoss), 64)
	if err != nil {
		return Config{}, fmt.Errorf("ping failure loss must be a number")
	}
	if cfg.Ping.WarnLoss < 0 || cfg.Ping.WarnLoss > 100 || cfg.Ping.FailLoss < 0 || cfg.Ping.FailLoss > 100 {
		return Config{}, fmt.Errorf("ping loss thresholds must be between 0 and 100")
	}
	if cfg.Ping.FailLoss > 0 && cfg.Ping.WarnLoss > cfg.Ping.FailLoss {
		return Config{}, fmt.Errorf("ping warning loss threshold must not exceed failure threshold")
	}
	cfg.Ping.Privileged = true

	// targets
	cfg.TCPTargets = filterEmpty(v.TCPTargets)
	cfg.PingTargets = filterEmpty(v.PingTargets)

	// resolvers (tags "name=address")
	if len(v.Resolvers) > 0 {
		out := make([]check.ResolverSpec, 0, len(v.Resolvers))
		for _, tag := range v.Resolvers {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			name, addr := tag, ""
			if i := strings.Index(tag, "="); i >= 0 {
				name = strings.TrimSpace(tag[:i])
				addr = strings.TrimSpace(tag[i+1:])
			}
			if name == "" {
				name = addr
			}
			out = append(out, check.ResolverSpec{Name: name, Address: addr})
		}
		cfg.Resolvers = out
	} else {
		cfg.Resolvers = nil
	}

	// queries
	var qs []check.QuerySpec
	for _, qv := range v.Queries {
		if strings.TrimSpace(qv.Name) == "" {
			continue
		}
		outcome := check.ExpectedSuccess
		expect := strings.ToUpper(strings.TrimSpace(qv.Expect))
		if expect == "NXDOMAIN" {
			outcome = check.ExpectedNXDomain
		} else if expect != "" && expect != "SUCCESS" {
			return Config{}, fmt.Errorf("DNS query %q has unknown expected outcome %q", qv.Name, qv.Expect)
		}
		var expected []string
		if outcome != check.ExpectedNXDomain {
			for _, a := range strings.Split(qv.Expected, ",") {
				a = strings.TrimSpace(a)
				if a != "" {
					expected = append(expected, a)
				}
			}
		} else if strings.TrimSpace(qv.Expected) != "" {
			return Config{}, fmt.Errorf("DNS query %q cannot have expected IPs with NXDOMAIN", qv.Name)
		}
		qs = append(qs, check.QuerySpec{
			Name:              qv.Name,
			ExpectedOutcome:   outcome,
			ExpectedAddresses: expected,
		})
	}
	if len(qs) > 0 {
		cfg.Queries = qs
	} else {
		cfg.Queries = nil
	}

	cfg.OutputPath = v.OutputResults
	cfg.SummaryPath = v.OutputSummary

	return cfg, nil
}

func splitLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}

func filterEmpty(in []string) []string {
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// SaveConfig writes cfg to path as YAML.
func SaveConfig(path string, cfg Config) error {
	fc := toFileConfig(cfg)
	b, err := yaml.Marshal(fc)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

func toFileConfig(cfg Config) fileConfig {
	fc := fileConfig{}
	if cfg.Timeout > 0 {
		fc.Timeout = cfg.Timeout.String()
	}
	c := cfg.Concurrency
	fc.Concurrency = &c
	fc.Monitor = &monitoringFile{DefaultInterval: cfg.Interval.String()}
	if len(cfg.Thresholds) > 0 {
		m := map[string]string{}
		for ct, d := range cfg.Thresholds {
			var name string
			switch ct {
			case domain.TCP:
				name = "tcp"
			case domain.DNS:
				name = "dns"
			case domain.Ping:
				name = "ping"
			}
			if name != "" {
				m[name] = d.String()
			}
		}
		fc.Thresholds = m
	}
	if cfg.OutputPath != "" || cfg.SummaryPath != "" {
		fc.Output = &outputFile{Results: cfg.OutputPath, Summary: cfg.SummaryPath}
	}
	// ping
	pc := cfg.Ping
	fc.Ping = &pingFile{
		Count:       &pc.Count,
		Interval:    pc.Interval.String(),
		Timeout:     pc.Timeout.String(),
		WarnLossPct: &pc.WarnLoss,
		FailLossPct: &pc.FailLoss,
		Privileged:  &pc.Privileged,
	}
	// targets
	fc.Targets = &targetsFile{
		TCP:  cfg.TCPTargets,
		Ping: cfg.PingTargets,
	}
	resolvers := make([]resolverFile, 0, len(cfg.Resolvers))
	for _, r := range cfg.Resolvers {
		resolvers = append(resolvers, resolverFile{Name: r.Name, Address: r.Address})
	}
	fc.Targets.Resolvers = resolvers
	// queries
	if len(cfg.Queries) > 0 {
		qs := make([]queryFile, 0, len(cfg.Queries))
		for _, q := range cfg.Queries {
			qf := queryFile{Name: q.Name}
			if q.ExpectedOutcome == check.ExpectedNXDomain {
				qf.Expect = "NXDOMAIN"
			} else {
				qf.Expected = q.ExpectedAddresses
			}
			qs = append(qs, qf)
		}
		fc.Queries = qs
	}
	return fc
}
