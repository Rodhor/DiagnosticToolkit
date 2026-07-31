package check

import (
	"context"

	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

type Check interface {
	Run(context.Context) []domain.CheckResult
}
