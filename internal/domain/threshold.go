package domain

import (
	"slices"
	"time"
)

type Thresholds map[CheckType]time.Duration

func ApplyThreshold(result CheckResult, t Thresholds) CheckResult {

	// If no threshold is present, return the results as-is
	if len(t) == 0 {
		return result
	}

	// If SkipLatencyThreshold is true, return the result as-is
	if result.SkipLatencyThreshold {
		return result
	}

	threshold, hasThreshold := t[result.CheckType]
	if result.Status == Pass && hasThreshold && threshold > 0 && result.Latency > threshold {
		result.Status = Warn
		result.Reason = ReasonHighLatency

		newMetrics := slices.Clone(result.Metrics)
		newMetrics = append(newMetrics, TextMetric(
			KeyWarnThreshold, "Latency Threshold", threshold.String(), false,
		))

		result.Metrics = newMetrics
	}

	return result

}

func ApplyThresholds(results []CheckResult, t Thresholds) []CheckResult {

	// If no threshold is present, return the results as-is
	if len(t) == 0 {
		return slices.Clone(results)
	}

	out := make([]CheckResult, 0, len(results))

	for _, res := range results {
		out = append(out, ApplyThreshold(res, t))
	}

	return out
}
