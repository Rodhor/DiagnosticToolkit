package check

import (
	"context"
	"fmt"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

type PingCheck struct {
	targets    []string
	count      int
	interval   time.Duration
	timeout    time.Duration
	warnLoss   float64
	failLoss   float64
	privileged bool
}

func NewPingCheck(
	targets []string,
	count int,
	interval time.Duration,
	timeout time.Duration,
	warnLoss float64,
	failLoss float64,
	privileged bool,
) *PingCheck {
	// Default number of pings
	if count <= 0 {
		count = 10
	}
	// Default interval
	if interval <= 0 {
		interval = 200 * time.Millisecond
	}

	// Floor timeout in cases where send schedule would exceed the maximum timeout
	minTimeout := time.Duration(count)*interval + 2*time.Second
	if timeout < minTimeout {
		timeout = minTimeout
	}

	return &PingCheck{
		targets:    targets,
		count:      count,
		interval:   interval,
		timeout:    timeout,
		warnLoss:   warnLoss,
		failLoss:   failLoss,
		privileged: privileged,
	}
}

func (c *PingCheck) Run(ctx context.Context) []domain.CheckResult {
	results := make([]domain.CheckResult, 0, len(c.targets))

	for _, target := range c.targets {
		res, err := c.pingTarget(ctx, target)
		// if context was cancelled, return results as is
		if err != nil {
			return results
		}
		results = append(results, res)
	}

	return results
}

func (c *PingCheck) pingTarget(ctx context.Context, target string) (domain.CheckResult, error) {
	start := time.Now()

	pinger, err := probing.NewPinger(target)
	if err != nil {
		detail := map[string]any{domain.KeyReason: domain.ReasonResolveFailure}
		return domain.NewCheckResult(
			domain.Ping, target, domain.Fail, 0, start, err, detail,
		), nil
	}

	pinger.Count = c.count
	pinger.Interval = c.interval
	pinger.Timeout = c.timeout
	pinger.SetPrivileged(c.privileged)
	runErr := pinger.RunWithContext(ctx)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return domain.CheckResult{}, ctxErr
	}

	if runErr != nil {
		return c.runErrorResult(target, start, runErr), nil
	}
	return c.statsResult(target, start, pinger.Statistics()), nil
}

func (c *PingCheck) statsResult(target string, start time.Time, stats *probing.Statistics) domain.CheckResult {
	detail := map[string]any{
		domain.KeyPacketsSent:     stats.PacketsSent,
		domain.KeyPacketsReceived: stats.PacketsRecv,
		domain.KeyPacketLoss:      stats.PacketLoss,
	}
	if stats.IPAddr != nil {
		detail[domain.KeyResolvedIP] = stats.IPAddr.String()
	}

	if stats.PacketsRecv == 0 {
		detail[domain.KeyReason] = domain.ReasonTotalLoss
		return domain.NewCheckResult(
			domain.Ping, target, domain.Fail, 0, start, fmt.Errorf("no reply to %d packets", stats.PacketsSent),
			detail,
		)
	}

	detail[domain.KeyMinRTT] = stats.MinRtt.String()
	detail[domain.KeyMaxRTT] = stats.MaxRtt.String()
	detail[domain.KeyAvgRTT] = stats.AvgRtt.String()
	detail[domain.KeyJitter] = stats.StdDevRtt.String()

	status := domain.Pass
	switch {
	case c.warnLoss > 0 && stats.PacketLoss >= c.warnLoss:
		status = domain.Warn
		detail[domain.KeyReason] = domain.ReasonPacketLoss
		detail[domain.KeyLossThreshold] = c.warnLoss
	case c.failLoss > 0 && stats.PacketLoss >= c.failLoss:
		status = domain.Fail
		detail[domain.KeyReason] = domain.ReasonPacketLoss
		detail[domain.KeyLossThreshold] = c.failLoss
	}

	return domain.NewCheckResult(
		domain.Ping, target, status, stats.AvgRtt, start, nil, detail,
	)

}

func (c *PingCheck) runErrorResult(target string, start time.Time, err error) domain.CheckResult {
	detail := map[string]any{}

	if isPermissionError(err) {
		detail[domain.KeyReason] = domain.ReasonNoPermission
		return domain.NewCheckResult(
			domain.Ping, target, domain.Unknown, 0, start, fmt.Errorf("ICMP socket denied: %w "+
				"Linux: Check `sysctl net.ipv4.ping_group_range`; "+
				"or run privileged", err),
			detail,
		)
	}

	detail[domain.KeyReason] = ClassifyError(err)
	return domain.NewCheckResult(
		domain.Ping, target, domain.Fail, 0, start, err, detail,
	)
}
