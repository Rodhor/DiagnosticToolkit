package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/check"
	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

func (fc fileConfig) merge(base Config) (Config, error) {
	cfg := base

	var err error
	if cfg.Timeout, err = parseDuration(fc.Timeout, base.Timeout); err != nil {
		return Config{}, fmt.Errorf("timeout: %w", err)
	}
	if fc.Concurrency != nil {
		if *fc.Concurrency < 0 {
			return Config{}, fmt.Errorf("concurrency must be non-negative: %d", *fc.Concurrency)
		}
		cfg.Concurrency = *fc.Concurrency
	}
	if fc.Monitor != nil {
		if cfg.Interval, err = parseDuration(
			fc.Monitor.DefaultInterval, base.Interval); err != nil {
			return Config{}, fmt.Errorf("monitor.default_interval: %w", err)
		}
	}

	if fc.Output != nil {
		if fc.Output.Results != "" {
			cfg.OutputPath = fc.Output.Results
		}
		if fc.Output.Summary != "" {
			cfg.SummaryPath = fc.Output.Summary
		}
	}

	if fc.Thresholds != nil {
		if cfg.Thresholds, err = convertThresholds(fc.Thresholds); err != nil {
			return Config{}, err
		}
	}

	if fc.Ping != nil {
		if cfg.Ping, err = fc.Ping.merge(base.Ping); err != nil {
			return Config{}, err
		}
	}

	if fc.Targets != nil {
		cfg.TCPTargets = fc.Targets.TCP
		cfg.PingTargets = fc.Targets.Ping
		cfg.Resolvers = convertResolvers(fc.Targets.Resolvers)
	}
	if fc.Queries != nil {
		if cfg.Queries, err = convertQueries(fc.Queries); err != nil {
			return Config{}, err
		}
	}
	return cfg, nil
}

func parseDuration(s string, fallback time.Duration) (time.Duration, error) {
	if s == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	if d < 0 {
		return 0, fmt.Errorf("duration must be non-negative: %q", s)
	}
	return d, nil
}

var checkTypesByName = map[string]domain.CheckType{
	"tcp":  domain.TCP,
	"dns":  domain.DNS,
	"ping": domain.Ping,
}

func convertThresholds(in map[string]string) (domain.Thresholds, error) {
	out := make(domain.Thresholds, len(in))

	for k, v := range in {
		ct, ok := checkTypesByName[strings.ToLower(k)]
		if !ok {
			return nil, fmt.Errorf(
				"thresholds: unknown check type %q (valid: tcp, dns, ping)", k)
		}
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("thresholds.%s: %w", k, err)
		}
		if d < 0 {
			return nil, fmt.Errorf("thresholds.%s: must be non-negative", k)
		}
		out[ct] = d
	}
	return out, nil
}
func (pf pingFile) merge(base check.PingConfig) (check.PingConfig, error) {
	out := base
	var err error

	if pf.Count != nil {
		if *pf.Count <= 0 {
			return out, errors.New("ping.count must be non-negative")
		}
		out.Count = *pf.Count
	}
	if out.Interval, err = parseDuration(pf.Interval, base.Interval); err != nil {
		return out, fmt.Errorf("ping.interval: %w", err)
	}
	if out.Timeout, err = parseDuration(pf.Timeout, base.Timeout); err != nil {
		return out, fmt.Errorf("ping.timeout: %w", err)
	}
	if pf.WarnLossPct != nil {
		out.WarnLoss = *pf.WarnLossPct
	}
	if pf.FailLossPct != nil {
		out.FailLoss = *pf.FailLossPct
	}
	if pf.Privileged != nil {
		out.Privileged = *pf.Privileged
	}

	if out.WarnLoss < 0 || out.WarnLoss > 100 ||
		out.FailLoss < 0 || out.FailLoss > 100 {
		return out, errors.New("ping loss thresholds must be between 0 and 100")
	}
	if out.FailLoss > 0 && out.WarnLoss > out.FailLoss {
		return out, fmt.Errorf(
			"ping.warn_loss_pct (%.0f) must not exceed fail_loss_pct (%.0f); "+
				"the warn threshold would never fire",
			out.WarnLoss, out.FailLoss)
	}

	return out, nil
}
func convertResolvers(in []resolverFile) []check.ResolverSpec {
	out := make([]check.ResolverSpec, 0, len(in))
	for _, r := range in {
		out = append(out, check.ResolverSpec{Name: r.Name, Address: r.Address})
	}
	return out
}

func convertQueries(in []queryFile) ([]check.QuerySpec, error) {
	out := make([]check.QuerySpec, 0, len(in))

	for i, q := range in {
		if q.Name == "" {
			return nil, fmt.Errorf("dns_queries[%d]: name is required", i)
		}

		outcome := check.ExpectedSuccess
		switch strings.ToUpper(q.Expect) {
		case "", "SUCCESS":
			// default
		case "NXDOMAIN":
			outcome = check.ExpectedNXDomain
		default:
			return nil, fmt.Errorf(
				"dns_queries[%d] (%s): unknown expect %q (valid: SUCCESS, NXDOMAIN)",
				i, q.Name, q.Expect)
		}

		if outcome == check.ExpectedNXDomain && len(q.Expected) > 0 {
			return nil, fmt.Errorf(
				"dns_queries[%d] (%s): expected addresses make no sense with "+
					"expect: NXDOMAIN", i, q.Name)
		}

		out = append(out, check.QuerySpec{
			Name:              q.Name,
			ExpectedOutcome:   outcome,
			ExpectedAddresses: q.Expected,
		})
	}
	return out, nil
}
