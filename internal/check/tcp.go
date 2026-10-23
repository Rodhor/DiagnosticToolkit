package check

import (
	"context"
	"net"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

type TCPCheck struct {
	targets []string
	timeout time.Duration
	dialer  *net.Dialer
}

func NewTCPCheck(targets []string, timeout time.Duration, dialer *net.Dialer) *TCPCheck {
	var d net.Dialer
	if dialer != nil {
		d = *dialer
	}
	d.Timeout = timeout
	return &TCPCheck{
		targets: targets,
		timeout: timeout,
		dialer:  &d,
	}
}

func (c *TCPCheck) Run(ctx context.Context) []domain.CheckResult {
	results := make([]domain.CheckResult, 0, len(c.targets))
	for _, target := range c.targets {
		// Break if context is cancelled
		if err := ctx.Err(); err != nil {
			break
		}

		// Returns false if context is cancelled
		res, err := c.checkTarget(ctx, target)
		if err != nil {
			break
		}
		results = append(results, res)
	}
	return results
}

func (c *TCPCheck) checkTarget(ctx context.Context, target string) (domain.CheckResult, error) {
	start := time.Now()
	conn, err := c.dialer.DialContext(ctx, "tcp", target)
	latency := time.Since(start)

	if ctx.Err() != nil {
		return domain.CheckResult{}, ctx.Err()
	}

	if err != nil {
		return domain.NewResultBuilder(domain.TCP, target).
			Status(domain.Fail).
			Latency(latency).
			Timestamp(start).
			Err(err).
			Reason(ClassifyError(err)).
			Build(), nil
	}

	remoteAddr := conn.RemoteAddr().String()
	_ = conn.Close()

	return domain.NewResultBuilder(domain.TCP, target).
		Status(domain.Pass).
		Latency(latency).
		Timestamp(start).
		Metric(domain.TextMetric(domain.KeyRemoteAddr, "Remote Endpoint", remoteAddr, false)).
		Build(), nil
}
