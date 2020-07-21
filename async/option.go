package async

// Options represents execution configurations. See Option functions for details on how
// this struct can be configured.
type Options struct {
	args         []interface{}
	goroutineNum int
}

// newOptions returns *Options with the given list of Option applied to it.
func newOptions(options ...Option) *Options {
	opts := new(Options)

	for _, option := range options {
		option(opts)
	}

	return opts
}

// Option represents an Options configuration function.
type Option func(*Options)

// WithArgs returns an Option for execution functions with the given args.
// How args will be used in execution depend on the type of async function
// this Option is passed to.
func WithArgs(args ...interface{}) Option {
	return func(opts *Options) { opts.args = args }
}

// WithGoroutineNum returns an Option for controlling the number of goroutines
// spawned by async functions.
func WithGoroutineNum(n int) Option {
	return func(opts *Options) { opts.goroutineNum = n }
}
