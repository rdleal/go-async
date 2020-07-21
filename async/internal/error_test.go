package internal

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestContextError_Error(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	e := NewContextError(ctx)

	if got, want := e.Error(), context.Canceled.Error(); !strings.Contains(got, want) {
		t.Errorf("Got %q; want it to contain %q.", got, want)
	}
}

func TestContextError_Unwrap(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	got := NewContextError(ctx)

	if want := context.DeadlineExceeded; !errors.Is(got, want) {
		t.Errorf("Got unwrapped error: %v; want %v.", got, want)
	}
}
