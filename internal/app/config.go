package app

import (
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/check"
	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

type derived struct {
	Gateway           string
	InternalResolvers []check.ResolverSpec
}

type Config struct {
	TCPTargets  []string
	PingTargets []string
	Resolvers   []check.ResolverSpec
	Queries     []check.QuerySpec
	Thresholds  domain.Thresholds
	Timeout     time.Duration
	Ping        check.PingConfig
	Concurrency int
	Interval    time.Duration
	Monitor     bool
	OutputPath  string
	SummaryPath string
}

func Default() Config {
	return Config{
		Timeout:     3 * time.Second,
		Concurrency: 4,
		Interval:    30 * time.Second,
		Thresholds: domain.Thresholds{
			domain.TCP:  300 * time.Millisecond,
			domain.DNS:  200 * time.Millisecond,
			domain.Ping: 150 * time.Millisecond,
		},
		Ping: check.PingConfig{
			Count:      10,
			Interval:   200 * time.Millisecond,
			Timeout:    5 * time.Second,
			WarnLoss:   1.0,
			FailLoss:   50.0,
			Privileged: true,
		},
	}
}
