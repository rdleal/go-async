package async

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestFuncResult_Index(t *testing.T) {
	testCases := []struct {
		name   string
		result FuncResult
		wantOK bool
		wantID interface{}
	}{
		{
			name:   "ValidIndex",
			result: FuncResult{id: 1},
			wantOK: true,
			wantID: 1,
		},
		{
			name:   "InvalidIndex",
			result: FuncResult{id: "funcname"},
			wantOK: false,
			wantID: 0,
		},
		{
			name:   "NilID",
			result: FuncResult{},
			wantOK: false,
			wantID: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotID, gotOK := tc.result.Index()

			if gotOK != tc.wantOK {
				t.Fatalf("FuncResult.Index(): got ok value: %t; want %t.", gotOK, tc.wantOK)
			}

			if !reflect.DeepEqual(gotID, tc.wantID) {
				t.Errorf("FuncResult.Index(): got id value: %v; want %v.", gotID, tc.wantID)
			}
		})
	}
}

func TestFuncResult_Name(t *testing.T) {
	testCases := []struct {
		name   string
		result FuncResult
		wantOK bool
		wantID interface{}
	}{
		{
			name:   "ValidName",
			result: FuncResult{id: "funcname"},
			wantOK: true,
			wantID: "funcname",
		},
		{
			name:   "InvalidName",
			result: FuncResult{id: 0},
			wantOK: false,
			wantID: "",
		},
		{
			name:   "NilID",
			result: FuncResult{},
			wantOK: false,
			wantID: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotID, gotOK := tc.result.Name()

			if gotOK != tc.wantOK {
				t.Fatalf("FuncResult.Name(): got ok value: %t; want %t.", gotOK, tc.wantOK)
			}

			if !reflect.DeepEqual(gotID, tc.wantID) {
				t.Errorf("FuncResult.Name(): got id value: %v; want %v.", gotID, tc.wantID)
			}
		})
	}
}

var noopFn = func() {}

type contextFunc func() (context.Context, context.CancelFunc)

var ctxBackground = func() (context.Context, context.CancelFunc) {
	return context.Background(), noopFn
}

var ctxWithValue = func(parent context.Context, key, value interface{}) func() (context.Context, context.CancelFunc) {
	return func() (context.Context, context.CancelFunc) {
		return context.WithValue(parent, key, value), noopFn
	}
}

var ctxCanceled = func() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return ctx, cancel
}

var ctxWithTimeout = func(parent context.Context, timeout time.Duration) func() (context.Context, context.CancelFunc) {
	return func() (context.Context, context.CancelFunc) {
		return context.WithTimeout(parent, timeout)
	}
}

func equalValues(x, y Values) bool {
	if len(x) != len(y) {
		return false
	}

	for i, vx := range x {
		if vx.Interface() != y[i].Interface() {
			return false
		}
	}

	return true
}

var errFunc = errors.New("func error")

func TestWaterfall(t *testing.T) {
	testCases := []struct {
		name       string
		ctx        contextFunc
		fns        Funcs
		options    []Option
		wantErr    error
		wantValues Values
	}{
		{
			name: "Exec",
			ctx:  ctxBackground,
			fns: Funcs{
				func(s string) string {
					return s + "Go"
				},
				func(s string) string {
					return s + "pher"
				},
				func(s string) string {
					return s + "!!"
				},
			},
			options: []Option{
				WithArgs("Hello, "),
			},
			wantValues: Values{reflect.ValueOf("Hello, Gopher!!")},
		},
		{
			name: "ExecWithContext",
			ctx:  ctxWithValue(context.Background(), "test", "value from context"),
			fns: Funcs{
				func(ctx context.Context) string {
					return ctx.Value("test").(string)
				},
			},
			wantValues: Values{reflect.ValueOf("value from context")},
		},
		{
			name: "ExecWithParamContext",
			ctx:  ctxBackground,
			fns: Funcs{
				func(ctx context.Context) string {
					return ctx.Value("test").(string)
				},
			},
			options: []Option{
				WithArgs(context.WithValue(context.Background(), "test", "value from parameter context")),
			},
			wantValues: Values{reflect.ValueOf("value from parameter context")},
		},
		{
			name: "ExecWithReturnedContext",
			ctx:  ctxBackground,
			fns: Funcs{
				func() context.Context {
					return context.WithValue(context.Background(), "test", "value from returned context")
				},
				func(ctx context.Context) string {
					return ctx.Value("test").(string)
				},
			},
			wantValues: Values{reflect.ValueOf("value from returned context")},
		},
		{
			name: "ExecWithReturnedError",
			ctx:  ctxBackground,
			fns: Funcs{
				func() (string, error) {
					return "nil-error", nil
				},
				func(string) (string, error) {
					return "", errFunc
				},
				func(string) string {
					return "never-called"
				},
			},
			wantErr:    errFunc,
			wantValues: nil,
		},
		{
			name: "ExecWithPanic",
			ctx:  ctxBackground,
			fns: Funcs{
				func() { panic("function panicked") },
			},
			wantErr:    ErrPanicked,
			wantValues: nil,
		},
		{
			name: "ExecWithContextDone",
			ctx:  ctxCanceled,
			fns: Funcs{
				func() { time.Sleep(1 * time.Millisecond) },
			},
			wantErr:    context.Canceled,
			wantValues: nil,
		},
		{
			name: "ExecWithContextTimeout",
			ctx:  ctxWithTimeout(context.Background(), 5*time.Millisecond),
			fns: Funcs{
				func() { time.Sleep(50 * time.Millisecond) },
				func() string { return "unwanted function call" },
			},
			wantErr:    context.DeadlineExceeded,
			wantValues: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := tc.ctx()
			defer cancel()

			fnc, err := Waterfall(ctx, tc.fns, tc.options...)
			if err != nil {
				t.Fatalf("Waterfall(%+v, %+v) = got error: %q.", tc.fns, tc.options, err)
			}

			var fn FuncResult
			select {
			case fn = <-fnc:
			case <-time.After(100 * time.Millisecond):
				t.Errorf("Got timed out waiting for function result.")
			}

			gotValues, err := fn.Returned()
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Got error %q; want error %q.", err, tc.wantErr)
			}

			if !equalValues(gotValues, tc.wantValues) {
				t.Errorf("Got returned values: %+v; want %+v", gotValues, tc.wantValues)
			}
		})
	}
}

func TestWaterfallError(t *testing.T) {
	testCases := []struct {
		name    string
		fns     Funcs
		wantErr error
	}{
		{
			name:    "InvalidFuncs",
			fns:     Funcs{"not a function"},
			wantErr: ErrInvalidFunc,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Waterfall(context.Background(), tc.fns)
			if err == nil {
				t.Fatalf("Waterfall(%+v) = got error: nil; want not nil.", tc.fns)
			}

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Got error: %v; want %v.", err, tc.wantErr)
			}
		})
	}
}

func TestConcurrent(t *testing.T) {
	testCases := []struct {
		name       string
		ctx        contextFunc
		fns        FuncMap
		options    []Option
		wantErr    error
		wantValues Values
	}{
		{
			name: "ExecFuncs",
			ctx:  ctxBackground,
			fns: FuncMap{
				"one": func(c chan<- int) {
					c <- 41
				},
				"two": func(c <-chan int) int {
					return 1 + <-c
				},
			},
			options: []Option{
				WithArgs(make(chan int)),
			},
			wantValues: Values{reflect.ValueOf(42)},
		},
		{
			name: "ExecFuncsMap",
			ctx:  ctxBackground,
			fns: FuncMap{
				"producer": func(c chan<- int) {
					c <- 41
				},
				"consumer": func(c <-chan int) int {
					return 1 + <-c
				},
			},
			options: []Option{
				WithArgs(make(chan int)),
			},
			wantValues: Values{reflect.ValueOf(42)},
		},
		{
			name: "ExecWithContext",
			ctx:  ctxWithValue(context.Background(), "test", "value from context"),
			fns: FuncMap{
				"fn": func(ctx context.Context) string {
					return ctx.Value("test").(string)
				},
			},
			wantValues: Values{reflect.ValueOf("value from context")},
		},
		{
			name: "ExecWithParamContext",
			ctx:  ctxBackground,
			fns: FuncMap{
				"fn": func(ctx context.Context) string {
					return ctx.Value("test").(string)
				},
			},
			options: []Option{
				WithArgs(context.WithValue(context.Background(), "test", "value from parameter context")),
			},
			wantValues: Values{reflect.ValueOf("value from parameter context")},
		},
		{
			name: "ExecWithError",
			ctx:  ctxBackground,
			fns: FuncMap{
				"fn": func() error {
					return errFunc
				},
			},
			wantErr: errFunc,
		},
		{
			name: "ExecWithPanic",
			ctx:  ctxBackground,
			fns: FuncMap{
				"fn": func() { panic("function panicked") },
			},
			wantErr: ErrPanicked,
		},
		{
			name: "ExecWithContextDone",
			ctx:  ctxCanceled,
			fns: FuncMap{
				"fn": func() {},
			},
			wantErr: context.Canceled,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := tc.ctx()
			defer cancel()

			fnc, err := Concurrent(ctx, tc.fns, tc.options...)
			if err != nil {
				t.Fatalf("Concurrent(%+v, %+v) = got error: %q.", tc.fns, tc.options, err)
			}

			var (
				gotValues Values
				gotErr    error
			)

			timeout := time.After(100 * time.Millisecond)

		loop:
			for {
				select {
				case fn, ok := <-fnc:
					if !ok {
						break loop
					}

					var v Values
					v, gotErr = fn.Returned()

					gotValues = append(gotValues, v...)
				case <-timeout:
					t.Fatal("Got timed out waiting for Concurrent function result.")
				}
			}

			if !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("Got error %q; want error %q.", gotErr, tc.wantErr)
			}

			if !equalValues(gotValues, tc.wantValues) {
				t.Errorf("Got returned values: %v; want %v", gotValues, tc.wantValues)
			}

		})
	}
}

func TestConcurrent_WithGoroutinesNum(t *testing.T) {
	testCases := []struct {
		name           string
		fns            FuncMap
		options        []Option
		wantGoroutines int32
	}{
		{
			name: "FuncMapLen",
			fns: FuncMap{
				"fn": func() {},
			},
			wantGoroutines: 1,
		},
		{
			name: "Arbitrary",
			fns: FuncMap{
				"fn0": func() {},
				"fn1": func() {},
				"fn2": func() {},
			},
			options: []Option{
				WithGoroutineNum(2),
			},
			wantGoroutines: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := atomic.LoadInt32(&goroutinesCount)

			fnc, err := Concurrent(context.Background(), tc.fns, tc.options...)
			if err != nil {
				t.Fatalf("Concurrent() = got error: %q.", err)
			}

			timeout := time.After(100 * time.Millisecond)

		loop:
			for {
				select {
				case _, ok := <-fnc:
					if !ok {
						break loop
					}
				case <-timeout:
					t.Fatal("Got timed out waiting for Concurrent function result.")
				}
			}

			now := atomic.LoadInt32(&goroutinesCount)
			if want := g + tc.wantGoroutines; now != want {
				t.Fatalf("Got %d goroutines created, want %d", now-g, tc.wantGoroutines)
			}
		})
	}
}

func TestConcurrentError(t *testing.T) {
	testCases := []struct {
		name    string
		fns     FuncMap
		wantErr error
	}{
		{
			name: "InvalidFuncsMap",
			fns: FuncMap{
				"NotAFunc": "not a function",
			},
			wantErr: ErrInvalidFunc,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Concurrent(context.Background(), tc.fns)
			if err == nil {
				t.Fatalf("Concurrent(%+v) = got error: nil; want not nil.", tc.fns)
			}

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Got error: %v; want %v.", err, tc.wantErr)
			}
		})
	}
}

func TestAuto(t *testing.T) {
	testCases := []struct {
		name       string
		ctx        contextFunc
		fns        FuncMap
		options    []Option
		wantErr    error
		wantValues Values
	}{
		{
			name: "ExecWithoutDependencies",
			ctx:  ctxBackground,
			fns: FuncMap{
				"leaf": func(i int) int {
					return i + 2
				},
			},
			options: []Option{
				WithArgs(40),
			},
			wantValues: Values{reflect.ValueOf(42)},
		},
		{
			name: "ExecWithDependencies",
			ctx:  ctxBackground,
			fns: FuncMap{
				"resolveD": func() int {
					return 42
				},
				"resolveC": func() int {
					return 30
				},
				"resolveB": DependsOn("resolveD", "resolveC").To(func(i, j int) int {
					return i + j
				}),
				"resolveA": DependsOn("resolveC", "resolveB").To(func(i, j int) int {
					return j - i
				}),
			},
			wantValues: Values{reflect.ValueOf(42)},
		},
		{
			name: "ExecWithContext",
			ctx:  ctxWithValue(context.Background(), "test", "value from context"),
			fns: FuncMap{
				"func": func(ctx context.Context) string {
					return ctx.Value("test").(string)
				},
			},
			wantValues: Values{reflect.ValueOf("value from context")},
		},
		{
			name: "ExecWithParamContext",
			ctx:  ctxBackground,
			fns: FuncMap{
				"func": func(ctx context.Context) string {
					return ctx.Value("test").(string)
				},
			},
			options: []Option{
				WithArgs(context.WithValue(context.Background(), "test", "value from parameter context")),
			},
			wantValues: Values{reflect.ValueOf("value from parameter context")},
		},
		{
			name: "ExecWithError",
			ctx:  ctxBackground,
			fns: FuncMap{
				"func": func() error {
					return errFunc
				},
				"dependentA": DependsOn("func").To(func() {
					t.Error("dependentA should never be called")
				}),
				"dependentB": DependsOn("dependentA").To(func() {
					t.Error("dependentB Should never be called")
				}),
			},
			wantErr: errFunc,
		},
		{
			name: "ExecWithDependentFuncError",
			ctx:  ctxBackground,
			fns: FuncMap{
				"func": func() {},
				"dependent": DependsOn("func").To(func() error {
					return errFunc
				}),
			},
			wantErr: errFunc,
		},
		{
			name: "ExecWithContextDone",
			ctx:  ctxCanceled,
			fns: FuncMap{
				"func":  func() {},
				"func2": func() {},
			},
			wantErr: context.Canceled,
		},
		{
			name: "ExecWithContextTimeout",
			ctx:  ctxWithTimeout(context.Background(), 5*time.Millisecond),
			fns: FuncMap{
				"func": func() { time.Sleep(50 * time.Millisecond) },
				"dependent": DependsOn("func").To(func() {
					t.Error("Got called; want it not to be called.")
				}),
			},
			wantErr: context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := tc.ctx()
			defer cancel()

			fnc, err := Auto(ctx, tc.fns, tc.options...)
			if err != nil {
				t.Fatalf("Auto(%+v, %+v) = got error: %q.", tc.fns, tc.options, err)
			}

			var (
				gotValues Values
				gotErr    error
			)

			timeout := time.After(500 * time.Millisecond)

		loop:
			for {
				select {
				case fn, ok := <-fnc:
					if !ok {
						break loop
					}

					var v Values
					v, gotErr = fn.Returned()

					gotValues = append(gotValues, v...)
				case <-timeout:
					t.Fatal("Got timed out waiting for Auto function result.")
				}
			}

			if !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("Got error %q; want error %q.", gotErr, tc.wantErr)
			}

			if !equalValues(gotValues, tc.wantValues) {
				t.Errorf("Got returned values: %v; want %v", gotValues, tc.wantValues)
			}
		})
	}
}

func TestAuto_WithGoroutinesNum(t *testing.T) {
	testCases := []struct {
		name           string
		fns            FuncMap
		options        []Option
		wantGoroutines int32
	}{
		{
			name: "FuncMapLen",
			fns: FuncMap{
				"fn": func() {},
			},
			wantGoroutines: 1,
		},
		{
			name: "Arbitrary",
			fns: FuncMap{
				"fn0": func() {},
				"fn1": func() {},
				"fn2": func() {},
			},
			options: []Option{
				WithGoroutineNum(2),
			},
			wantGoroutines: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := atomic.LoadInt32(&goroutinesCount)

			fnc, err := Auto(context.Background(), tc.fns, tc.options...)
			if err != nil {
				t.Fatalf("Auto() = got error: %q.", err)
			}

			timeout := time.After(100 * time.Millisecond)

		loop:
			for {
				select {
				case _, ok := <-fnc:
					if !ok {
						break loop
					}
				case <-timeout:
					t.Fatal("Got timed out waiting for Auto function result.")
				}
			}

			now := atomic.LoadInt32(&goroutinesCount)
			if want := g + tc.wantGoroutines; now != want {
				t.Fatalf("Got %d goroutines created, want %d", now-g, tc.wantGoroutines)
			}
		})
	}
}

func TestAutoError(t *testing.T) {
	testCases := []struct {
		name    string
		fns     FuncMap
		wantErr error
	}{
		{
			name: "InvalidFuncsMap",
			fns: FuncMap{
				"NotAFunc": "not a function",
			},
			wantErr: ErrInvalidFunc,
		},
		{
			name: "InvalidDependentFunc",
			fns: FuncMap{
				"NotAFunc": DependsOn().To("not a function"),
			},
			wantErr: ErrInvalidFunc,
		},
		{
			name: "InvalidDependencyGraph",
			fns: FuncMap{
				"func": DependsOn("func").To(func() {}),
			},
			wantErr: ErrDependencyGraph,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Auto(context.Background(), tc.fns)
			if err == nil {
				t.Fatalf("Auto(%+v) = got error: nil; want not nil.", tc.fns)
			}

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Got error: %v; want %v.", err, tc.wantErr)
			}
		})
	}
}

func makeFuncs(fn interface{}, n int) Funcs {
	fns := make(Funcs, n)
	for i := 0; i < n; i++ {
		fns[i] = fn
	}
	return fns
}

func makeFuncsMap(fn interface{}, n int) FuncMap {
	fns := make(FuncMap)
	for i := 0; i < n; i++ {
		fns[fmt.Sprintf("func-%d", i)] = fn
	}
	return fns
}

func BenchmarkWaterfall(b *testing.B) {
	testCases := []struct {
		name string
		ctx  context.Context
		fns  Funcs
		opts []Option
	}{
		{
			name: "10Funcs",
			ctx:  context.Background(),
			fns:  makeFuncs(noopFn, 10),
		},
		{
			name: "1000Funcs",
			ctx:  context.Background(),
			fns:  makeFuncs(noopFn, 1000),
		},
		{
			name: "ContextAndError",
			ctx:  context.Background(),
			fns:  makeFuncs(func(ctx context.Context) error { return nil }, 1000),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				fnc, err := Waterfall(tc.ctx, tc.fns, tc.opts...)
				if err != nil {
					b.Fatalf("Waterfall() = got error: %q; want nil.", err)
				}

				for range fnc {
				}
			}
		})
	}
}

func BenchmarkConcurrent(b *testing.B) {
	testCases := []struct {
		name    string
		fns     FuncMap
		options []Option
	}{
		{
			name: "10Funcs",
			fns:  makeFuncsMap(noopFn, 10),
		},
		{
			name: "1000Funcs",
			fns:  makeFuncsMap(noopFn, 1000),
			options: []Option{
				WithGoroutineNum(10),
			},
		},
		{
			name: "ContextAndError",
			fns:  makeFuncsMap(func(ctx context.Context) error { return nil }, 1000),
			options: []Option{
				WithGoroutineNum(10),
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				fnc, err := Concurrent(context.Background(), tc.fns, tc.options...)
				if err != nil {
					b.Fatalf("Concurrent() = got error: %q; want nil.", err)
				}

				for range fnc {
				}
			}
		})
	}
}

func makeDependncyGraph(fn interface{}, n int) FuncMap {
	fns := make(FuncMap)
	var prev string
	for i := 0; i < n; i++ {
		funcName := fmt.Sprintf("func-%d", i)
		if i%10 == 0 {
			fns[funcName] = noopFn
		} else {
			fns[funcName] = DependsOn(prev).To(noopFn)
		}
		prev = funcName
	}

	return fns
}

func BenchmarkAuto(b *testing.B) {
	testCases := []struct {
		name    string
		fns     FuncMap
		options []Option
	}{
		{
			name: "10LeafFuncs",
			fns:  makeFuncsMap(noopFn, 10),
		},
		{
			name: "1000Funcs",
			fns:  makeDependncyGraph(noopFn, 1000),
		},
		{
			name: "ContextAndError",
			fns:  makeDependncyGraph(func(ctx context.Context) error { return nil }, 1000),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				fnc, err := Auto(context.Background(), tc.fns, tc.options...)
				if err != nil {
					b.Fatalf("Auto() = got error: %q; want nil.", err)
				}

				for range fnc {
				}
			}
		})
	}
}
