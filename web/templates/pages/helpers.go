package pages

import (
	"fmt"

	"github.com/rodhor/DiagnosticToolset/internal/app"
)

func fmtInt(n int) string {
	return fmt.Sprintf("%d", n)
}

func fmtFloat(f float64) string {
	return fmt.Sprintf("%g", f)
}

func qName(qs []app.QueryView, i int) string {
	if i < len(qs) {
		return qs[i].Name
	}
	return ""
}

func qExpected(qs []app.QueryView, i int) string {
	if i < len(qs) {
		return qs[i].Expected
	}
	return ""
}

func qExpect(qs []app.QueryView, i int) string {
	if i < len(qs) {
		return qs[i].Expect
	}
	return ""
}

func qRowCount(qs []app.QueryView) int {
	n := len(qs) + 1
	if n < 3 {
		n = 3
	}
	if n > 12 {
		n = 12
	}
	return n
}
