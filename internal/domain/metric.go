package domain

import (
	"fmt"
	"strconv"
	"time"
)

type Metric struct {
	Key     string
	Label   string
	Value   any
	Display string
	Primary bool
}

func DurationMetric(key string, label string, d time.Duration, primary bool) Metric {
	return Metric{
		Key:     key,
		Label:   label,
		Value:   d.Milliseconds(),
		Display: d.Round(time.Microsecond).String(),
		Primary: primary,
	}
}

func PercentMetric(key string, label string, pct float64, primary bool) Metric {
	return Metric{
		Key:     key,
		Label:   label,
		Value:   pct,
		Display: fmt.Sprintf("%.0f%%", pct),
		Primary: primary,
	}
}

func TextMetric(key string, label string, value string, primary bool) Metric {
	return Metric{
		Key:     key,
		Label:   label,
		Value:   value,
		Display: value,
		Primary: primary,
	}
}

func IntegerMetric(key string, label string, value int, primary bool) Metric {
	return Metric{
		Key:     key,
		Label:   label,
		Value:   value,
		Display: strconv.Itoa(value),
		Primary: primary,
	}
}

func FloatMetric(key string, label string, value float64, primary bool) Metric {
	return Metric{
		Key:     key,
		Label:   label,
		Value:   value,
		Display: fmt.Sprintf("%g", value),
		Primary: primary,
	}
}
