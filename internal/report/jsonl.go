package report

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

type JSONLReporter struct {
	w   io.Writer
	enc *json.Encoder
	mu  sync.Mutex
}

func NewJSONLReporter(w io.Writer) *JSONLReporter {
	return &JSONLReporter{
		w:   w,
		enc: json.NewEncoder(w),
	}
}

func (jr *JSONLReporter) WriteResult(result domain.CheckResult) error {
	jr.mu.Lock()
	defer jr.mu.Unlock()
	return jr.enc.Encode(result)
}
