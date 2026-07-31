package domain

import (
	"slices"
	"time"
)

type Thresholds map[CheckType]time.Duration

func ApplyThresholds(results []CheckResult, t Thresholds) []CheckResult {

	// If no threshold is present, return the results as-is
	if len(t) == 0 {
		return slices.Clone(results)
	}

	// Prepare the output slice
	out := make([]CheckResult, 0, len(results))

	for _, res := range results {
		r := res // Shallow copy

		threshold, hasThreshold := t[res.CheckType]
		if r.Status == Pass && hasThreshold && threshold > 0 && r.Latency > threshold {
			r.Status = Warn
			r.Reason = ReasonHighLatency

			newMetrics := slices.Clone(r.Metrics)
			newMetrics = append(newMetrics, TextMetric(
				KeyWarnThreshold, "Latency Threshold", threshold.String(), false,
			))

			r.Metrics = newMetrics
		}

		out = append(out, r)
	}

	return out
}
