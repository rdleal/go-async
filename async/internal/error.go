package internal

import (
	"context"
	"fmt"
)

// ContextError is sent to async functions output channel when a given Context is canceled or timed out
// before the execution flow is completed. This error will be sent as the last message in the output channel,
// so the user will receive all the results from functions that have started before the Context error,
// before being notified about such error.
type ContextError struct {
	ctx context.Context
}

func (e *ContextError) Error() string {
	return fmt.Sprintf("[async] context error: %s", e.ctx.Err())
}

// Unwrap unwraps the original Context error so the error chain can be inspected.
func (e *ContextError) Unwrap() error {
	return e.ctx.Err()
}

func NewContextError(ctx context.Context) *ContextError {
	return &ContextError{ctx}
}
