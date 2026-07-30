package check

import (
	"context"
	"errors"
	"net"
	"syscall"
	"time"
)

const (
	KeyReason        = "reason"
	KeyRemoteAddr    = "remote_addr"
	KeyWarnThreshold = "warn_threshold"
)

const (
	ReasonTimeout     = "timeout"
	ReasonRefused     = "refused"
	ReasonUnknown     = "unknown"
	ReasonHighLatency = "high_latency"
	ReasonDNSFailure  = "dns_failure"
)

type TCPCheck struct {
	targets       []string
	timeout       time.Duration
	warnThreshold time.Duration
	dialer        *net.Dialer
}

func NewTCPCheck(targets []string, timeout time.Duration, warnThreshold time.Duration, dialer *net.Dialer) *TCPCheck {
	var d net.Dialer
	if dialer != nil {
		d = *dialer
	}
	d.Timeout = timeout
	return &TCPCheck{
		targets:       targets,
		timeout:       timeout,
		warnThreshold: warnThreshold,
		dialer:        &d,
	}
}

func (c *TCPCheck) Run(ctx context.Context) []Result {
	results := make([]Result, 0, len(c.targets))
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

func (c *TCPCheck) checkTarget(ctx context.Context, target string) (Result, bool) {
	start := time.Now()
	conn, err := c.dialer.DialContext(ctx, "tcp", target)
	latency := time.Since(start)

	if ctx.Err() != nil {
		return Result{}, false
	}

	if err != nil {
		return c.newResult(target, Fail, latency, start, err, map[string]any{KeyReason: classifyError(err)}), true
	}

	remoteAddr := conn.RemoteAddr().String()
	_ = conn.Close()

	status := Pass
	detail := map[string]any{
		KeyRemoteAddr: remoteAddr,
	}
	if c.warnThreshold > 0 && latency > c.warnThreshold {
		status = Warn
		detail[KeyReason] = ReasonHighLatency
		detail[KeyWarnThreshold] = c.warnThreshold.String()
	}
	return c.newResult(target, status, latency, start, nil, detail), true

}

func (c *TCPCheck) newResult(
	target string,
	status ResultStatus,
	latency time.Duration,
	timestamp time.Time,
	err error,
	detail map[string]any,
) Result {
	var errStr string
	if err != nil {
		errStr = err.Error()
	}
	return Result{
		CheckType: TCP,
		Target:    target,
		Status:    status,
		Latency:   latency,
		Err:       errStr,
		Timestamp: timestamp,
		Detail:    detail,
	}
}

func classifyError(err error) string {
	if errors.Is(err, syscall.ECONNREFUSED) {
		return ReasonRefused
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return ReasonDNSFailure
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return ReasonTimeout
	}
	return ReasonUnknown
}
