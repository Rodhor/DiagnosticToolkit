package platform

import "context"

// FakeProvider is a Provider for tests of downstream consumers.
type FakeProvider struct {
	Result Snapshot
	Err    error
}

var _ Provider = (*FakeProvider)(nil)

func (f *FakeProvider) Snapshot(context.Context) (Snapshot, error) {
	return f.Result, f.Err
}
