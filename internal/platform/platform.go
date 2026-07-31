package platform

import (
	"context"
	"errors"
	"fmt"
)

var ErrUnsupported = errors.New("platform provider not implemented for this OS")

type Provider interface {
	// Snapshot returns the current network environment. A non-nil error
	// means nothing useful could be obtained at all; partial failures
	// (unreadable lease file, missing route, permission denial) go in
	// Snapshot.Warnings and the partial snapshot is still returned.
	Snapshot(ctx context.Context) (Snapshot, error)
}

type unsupportedProvider struct {
	os string
}

func (u *unsupportedProvider) Snapshot(context.Context) (Snapshot, error) {
	return Snapshot{}, fmt.Errorf("%s: %w", u.os, ErrUnsupported)
}
