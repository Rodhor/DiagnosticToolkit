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
		res, ok := c.checkTarget(ctx, target)
		if !ok {
			break
		}
		results = append(results, res)
	}
	return results
}

func (c *TCPCheck) checkTarget(ctx context.Context, target string) (domain.CheckResult, bool) {
	start := time.Now()
	conn, err := c.dialer.DialContext(ctx, "tcp", target)
	latency := time.Since(start)

	if ctx.Err() != nil {
		return domain.CheckResult{}, false
	}

	if err != nil {
		return domain.NewCheckResult(domain.TCP, target, domain.Fail, latency, start, err, map[string]any{domain.KeyReason: ClassifyError(err)}), true
	}

	remoteAddr := conn.RemoteAddr().String()
	_ = conn.Close()

	status := domain.Pass
	detail := map[string]any{
		domain.KeyRemoteAddr: remoteAddr,
	}

	return domain.NewCheckResult(domain.TCP, target, status, latency, start, nil, detail), true

}
