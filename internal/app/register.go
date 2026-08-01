package app

import (
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/check"
	"github.com/rodhor/DiagnosticToolset/internal/domain"
	"github.com/rodhor/DiagnosticToolset/internal/runner"
)

const (
	intervalGateway = 5 * time.Second
	intervalPing    = 5 * time.Second
	intervalTCP     = 15 * time.Second
	intervalDNS     = 30 * time.Second
)

func registerChecks(cfg Config, d derived) []runner.RegisteredCheck {
	var out []runner.RegisteredCheck

	if d.Gateway != "" {
		out = append(out, runner.RegisteredCheck{
			CheckType: domain.Ping,
			Check:     check.NewPingCheck([]string{d.Gateway}, cfg.Ping),
			Name:      "Gateway",
			Category:  domain.CategorySystem,
			Interval:  intervalGateway,
			Serial:    false,
		})
	}

	if len(d.InternalResolvers) > 0 && len(cfg.Queries) > 0 {
		out = append(out, runner.RegisteredCheck{
			CheckType: domain.DNS,
			Check:     check.NewDNSCheck(d.InternalResolvers, cfg.Queries, cfg.Timeout),
			Name:      "Internal DNS",
			Category:  domain.CategorySystem,
			Interval:  intervalDNS,
		})
	}

	if len(cfg.Resolvers) > 0 && len(cfg.Queries) > 0 {
		out = append(out, runner.RegisteredCheck{
			CheckType: domain.DNS,
			Check:     check.NewDNSCheck(cfg.Resolvers, cfg.Queries, cfg.Timeout),
			Name:      "Public DNS",
			Category:  domain.CategoryInternet,
			Interval:  intervalDNS,
		})
	}

	if len(cfg.TCPTargets) > 0 {
		out = append(out, runner.RegisteredCheck{
			CheckType: domain.TCP,
			Check:     check.NewTCPCheck(cfg.TCPTargets, cfg.Timeout, nil),
			Name:      "Internet TCP",
			Category:  domain.CategoryInternet,
			Interval:  intervalTCP,
		})
	}

	if len(cfg.PingTargets) > 0 {
		out = append(out, runner.RegisteredCheck{
			CheckType: domain.Ping,
			Check:     check.NewPingCheck(cfg.PingTargets, cfg.Ping),
			Name:      "Internet Ping",
			Category:  domain.CategoryInternet,
			Interval:  intervalPing,
			Serial:    true,
		})
	}

	return out
}
