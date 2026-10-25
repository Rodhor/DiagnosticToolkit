package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

type JSONLReporter struct {
	enc    *json.Encoder
	closer io.Closer
	mu     sync.Mutex
}

func NewJSONLReporter(w io.Writer) *JSONLReporter {
	return &JSONLReporter{
		enc: json.NewEncoder(w),
	}
}

func (jr *JSONLReporter) WriteResult(result domain.CheckResult) error {
	jr.mu.Lock()
	defer jr.mu.Unlock()
	return jr.enc.Encode(result)
}

func NewJSONLFileReporter(path string) (*JSONLReporter, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("creating directory for jsonl output: %w", err)
		}
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("opening jsonl output: %w", err)
	}

	r := NewJSONLReporter(f)
	r.closer = f
	return r, nil
}

func (jr *JSONLReporter) Close() error {
	jr.mu.Lock()
	defer jr.mu.Unlock()
	if jr.closer == nil {
		return nil
	}
	return jr.closer.Close()
}
