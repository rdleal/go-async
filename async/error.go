package async

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/rdleal/go-async/async/internal"
)

// err is an implementation of error interface.
// It's a private string type so t can be created as constant,
// thus making them immutable.
// See https://dave.cheney.net/2016/04/07/constant-errors
type err string

func (e err) Error() string {
	return fmt.Sprintf("[async] %s", string(e))
}

// ErrDependencyGraph means that some dependency relationship has cycles.
const ErrDependencyGraph = err("invalid dependency graph")

// ErrInvalidFunc means a given function is not a valid one.
const ErrInvalidFunc = err("invalid function value")

// ErrPanicked means a Panic occurred during a function execution.
const ErrPanicked = err("panicked")

// IsContextError reports whether err is a Context error occurred before any async function completed its execution flow.
func IsContextError(err error) bool {
	var ctxErr *internal.ContextError
	return errors.As(err, &ctxErr)
}

// ctxErr returns a read-only error channel for receiving a context error and a function for notifying the errc once.
func ctxErr() (errc <-chan error, notifyErrOnce func(context.Context)) {
	errch := make(chan error, 1)

	errc = errch

	var once sync.Once
	notifyErrOnce = func(ctx context.Context) {
		once.Do(func() {
			errch <- internal.NewContextError(ctx)
		})
	}

	return
}
