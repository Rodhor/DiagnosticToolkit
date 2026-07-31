package domain

import (
	"maps"
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

			newDetail := make(map[string]any, len(r.Detail)+2)
			// Copy the existing detail and add the reason and threshold
			maps.Copy(newDetail, r.Detail)

			// Add reason and threshold to the detail
			newDetail[KeyReason] = ReasonHighLatency
			newDetail[KeyWarnThreshold] = threshold.String()

			// Update the result detail
			r.Detail = newDetail
		}

		out = append(out, r)
	}

	return out
}
