package server

import "sync"

type ringBuffer[T any] struct {
	mu    sync.Mutex
	buf   []T
	size  int
	head  int
	count int
}

func newRingBuffer[T any](size int) *ringBuffer[T] {
	if size <= 0 {
		size = 1
	}
	return &ringBuffer[T]{buf: make([]T, size), size: size}
}

func (r *ringBuffer[T]) Add(v T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.head] = v
	r.head = (r.head + 1) % r.size
	if r.count < r.size {
		r.count++
	}
}

func (r *ringBuffer[T]) All() []T {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]T, 0, r.count)
	start := (r.head - r.count + r.size) % r.size
	for i := 0; i < r.count; i++ {
		out = append(out, r.buf[(start+i)%r.size])
	}
	return out
}
