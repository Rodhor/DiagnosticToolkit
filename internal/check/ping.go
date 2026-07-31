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
		return domain.NewResultBuilder(
			domain.Ping, target).
			Status(domain.Fail).
			Timestamp(start).
			Reason(domain.ReasonResolveFailure).
			Err(err).
			Build(), nil
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
	metrics := []domain.Metric{
		domain.IntegerMetric(domain.KeyPacketsSent, "Packets Sent", stats.PacketsSent, false),
		domain.IntegerMetric(domain.KeyPacketsReceived, "Packets Received", stats.PacketsRecv, false),
		domain.PercentMetric(domain.KeyPacketLoss, "Packets Lost", stats.PacketLoss, true),
	}

	if stats.IPAddr != nil {
		metrics = append(metrics, domain.TextMetric(
			domain.KeyResolvedIP, "Resolved IP", stats.IPAddr.String(), false))
	}

	if stats.PacketsRecv == 0 {
		return domain.NewResultBuilder(domain.Ping, target).
			Status(domain.Fail).
			Timestamp(start).
			Reason(domain.ReasonTotalLoss).
			Err(fmt.Errorf("no reply to %d packets", stats.PacketsSent)).
			Metrics(metrics).
			Build()
	}

	metrics = append(metrics,
		domain.TextMetric(domain.KeyMinRTT, "Min RTT", stats.MinRtt.String(), false),
		domain.TextMetric(domain.KeyMaxRTT, "Max RTT", stats.MaxRtt.String(), false),
		domain.TextMetric(domain.KeyAvgRTT, "Average RTT", stats.AvgRtt.String(), false),
		domain.TextMetric(domain.KeyJitter, "Jitter", stats.StdDevRtt.String(), false),
	)
	status := domain.Pass
	reason := ""
	switch {
	case c.failLoss > 0 && stats.PacketLoss >= c.failLoss:
		status = domain.Fail
		reason = domain.ReasonPacketLoss
		metrics = append(metrics,
			domain.PercentMetric(domain.KeyLossThreshold, "Loss Threshold", c.failLoss, false),
		)
	case c.warnLoss > 0 && stats.PacketLoss >= c.warnLoss:
		status = domain.Warn
		reason = domain.ReasonPacketLoss
		metrics = append(metrics,
			domain.PercentMetric(domain.KeyLossThreshold, "Loss Threshold", c.warnLoss, false),
		)
	}

	return domain.NewResultBuilder(domain.Ping, target).
		Status(status).
		Reason(reason).
		Latency(stats.AvgRtt).
		Timestamp(start).
		Metrics(metrics).
		Build()

}

func (c *PingCheck) runErrorResult(target string, start time.Time, err error) domain.CheckResult {
	if isPermissionError(err) {
		var hint string
		if c.privileged {
			hint = "raw ICMP socket denied — grant the capability with " +
				"`sudo setcap cap_net_raw+ep <binary>`, or run unprivileged"
		} else {
			hint = "unprivileged ICMP denied — enable it with " +
				"`sudo sysctl -w net.ipv4.ping_group_range=\"0 2147483647\"`, " +
				"or run privileged"
		}
		return domain.NewResultBuilder(domain.Ping, target).
			Status(domain.Unknown).
			Timestamp(start).
			Reason(domain.ReasonNoPermission).
			Err(fmt.Errorf("%s: %w", hint, err)).
			Build()
	}
	return domain.NewResultBuilder(domain.Ping, target).
		Status(domain.Fail).
		Timestamp(start).
		Err(err).
		Reason(ClassifyError(err)).
		Build()
}
