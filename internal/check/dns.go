package check

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

type ExpectedOutcome string

const (
	ExpectedSuccess  ExpectedOutcome = "SUCCESS"
	ExpectedNXDomain ExpectedOutcome = "NXDOMAIN"
)

type ResolverSpec struct {
	Name    string
	Address string
}

type QuerySpec struct {
	Name              string
	ExpectedOutcome   ExpectedOutcome
	ExpectedAddresses []string
}

type namedResolver struct {
	name       string
	targetAdrr string
	resolver   *net.Resolver
}

type DNSCheck struct {
	resolvers []namedResolver
	queries   []QuerySpec
	timeout   time.Duration
}

func NewDNSCheck(resolvers []ResolverSpec, queries []QuerySpec, timeout time.Duration) *DNSCheck {
	namedResolvers := make([]namedResolver, 0, len(resolvers))

	for _, spec := range resolvers {
		// Empty address defaults to system OS resolver
		if spec.Address == "" {
			namedResolvers = append(namedResolvers, namedResolver{
				name:     spec.Name,
				resolver: net.DefaultResolver,
			})
			continue
		}

		addr := spec.Address
		if _, _, err := net.SplitHostPort(addr); err != nil {
			addr = net.JoinHostPort(addr, "53")
		}

		// Address is omitted to use the system OS resolver
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network string, _ string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		}

		namedResolvers = append(namedResolvers, namedResolver{
			name:       spec.Name,
			targetAdrr: addr,
			resolver:   r,
		})
	}

	return &DNSCheck{
		resolvers: namedResolvers,
		queries:   queries,
		timeout:   timeout,
	}
}

func (c *DNSCheck) Run(ctx context.Context) []domain.CheckResult {
	results := make([]domain.CheckResult, 0, len(c.resolvers)*len(c.queries))

	for _, r := range c.resolvers {
		for _, query := range c.queries {
			res, err := c.checkQuery(ctx, r, query)
			if err != nil {
				// Stop execution only if context was canceled/timed out
				return results
			}
			results = append(results, res)
		}
	}

	return results
}

func (c *DNSCheck) checkQuery(
	ctx context.Context,
	resolver namedResolver,
	query QuerySpec,
) (domain.CheckResult, error) {
	queryCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	targetLabel := fmt.Sprintf("%s via %s", query.Name, resolver.name)
	start := time.Now()

	ips, err := resolver.resolver.LookupHost(queryCtx, query.Name)
	latency := time.Since(start)

	if err != nil {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) && resolver.targetAdrr != "" {
			dnsErr.Server = resolver.targetAdrr

		}
	}

	// Context cancellation check
	if ctxErr := ctx.Err(); ctxErr != nil {
		return domain.CheckResult{}, ctxErr
	}

	detail := map[string]any{
		domain.KeyResolver:  resolver.name,
		domain.KeyQueryName: query.Name,
	}

	outcome := query.ExpectedOutcome
	if outcome == "" {
		outcome = ExpectedSuccess
	}

	if outcome == ExpectedNXDomain {
		if err != nil {
			reason := ClassifyError(err)
			if reason == domain.ReasonNXDomain {
				return domain.NewCheckResult(domain.DNS, targetLabel, domain.Pass, latency, start, nil, detail), nil
			}

			// Non-NXDOMAIN error (e.g. Timeout)
			detail[domain.KeyReason] = reason
			return domain.NewCheckResult(domain.DNS, targetLabel, domain.Fail, latency, start, err, detail), nil
		}

		slices.Sort(ips)

		// Lookup succeeded on non-existent record -> Hijack detected
		detail[domain.KeyReason] = domain.ReasonNXDomainHijacked
		detail[domain.KeyActualAddresses] = ips
		return domain.NewCheckResult(
			domain.DNS,
			targetLabel,
			domain.Fail,
			latency,
			start,
			fmt.Errorf("lookup succeeded for non-existent name (possible DNS hijack)"),
			detail,
		), nil
	}

	if err != nil {
		detail[domain.KeyReason] = ClassifyError(err)
		return domain.NewCheckResult(domain.DNS, targetLabel, domain.Fail, latency, start, err, detail), nil
	}

	slices.Sort(ips)
	detail[domain.KeyActualAddresses] = ips

	if len(query.ExpectedAddresses) > 0 {
		sortedExpected := make([]string, len(query.ExpectedAddresses))
		copy(sortedExpected, query.ExpectedAddresses)
		slices.Sort(sortedExpected)
		detail[domain.KeyExpectedAddresses] = sortedExpected

		if !containsAll(ips, sortedExpected) {
			detail[domain.KeyReason] = domain.ReasonUnexpectedAnswer
			return domain.NewCheckResult(
				domain.DNS,
				targetLabel,
				domain.Fail,
				latency,
				start,
				fmt.Errorf("expected addresses %v not found in answer %v", sortedExpected, ips),
				detail,
			), nil
		}
	}

	return domain.NewCheckResult(
		domain.DNS,
		targetLabel,
		domain.Pass,
		latency,
		start,
		nil,
		detail,
	), nil
}

func containsAll(actual []string, expected []string) bool {
	actualSet := make(map[string]bool, len(actual))
	for _, a := range actual {
		actualSet[a] = true
	}
	for _, e := range expected {
		if !actualSet[e] {
			return false
		}
	}
	return true
}
