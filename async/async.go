// Package async provides utility functions for controlling asynchronous flow.
//
// Execution and Context Awareness
//
// In case the Context passed to any async function is canceled or timed out, such function will stop
// the execution flow, send a last message with the Context error and close their output channel as soon as possible.
// Which means they will wait for all functions that have been called but haven't returned yet before
// sending the error message and closing their output channel.
//
// 	ctx, cancel := context.WithCancel(context.Background())
//
// 	fnc, err := async.Waterfall(ctx, async.Funcs{
// 		func() { fmt.Println("first executed") },
// 		func() {
// 			cancel()
// 			fmt.Println("second executed")
// 		},
// 		func() { fmt.Println("never executed") },
// 	})
//
// 	for fn := range fnc {
// 	 	if vals, err := fn.Returned(); async.IsContextError(err) {
// 	 		log.Fatal(err)
// 	 	}
// 	}
// 	// ...
//
// Errors
//
// In order for async to identify if an error occurred, the error must be the last or only return
// value of the function:
//
// 	fnc, err := async.Concurrent(context.Background(), async.FuncMap{
// 		"failed": func() (int, error) {
// 			return 0, errors.New("function failed") // fnc will receive a failed result.
// 		},
// 		"also-failed": func() error {
// 			return errors.New("function failed too") // fnc will also receive a failed result.
// 		},
// 		"not-failed": func() (error, bool) {
// 			return errors.New("some error"), false // fnc will receive a successful result.
// 		},
// 	})
// 	// ...
//
// Panicking
//
// If a function panics, async will recover and send a message to output channel with the ErrPanicked error.
package async

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/rdleal/go-async/async/internal"
)

type function struct {
	ID    interface{}
	Value reflect.Value
}

// Funcs is a slice for defining a list of functions to be executed.
type Funcs []interface{}

func (f Funcs) funcs() ([]function, error) {
	var fns []function
	for i, fn := range f {
		v := reflect.ValueOf(fn)
		if !isFunc(v) {
			return nil, fmt.Errorf("%w at #%d: %v", ErrInvalidFunc, i, v)
		}

		fns = append(fns, function{i, v})
	}

	return fns, nil
}

// FuncMap is the type for mapping from names to functions to be executed.
type FuncMap map[string]interface{}

func (fm FuncMap) funcs() ([]function, error) {
	var fns []function
	for k, fn := range fm {
		v := reflect.ValueOf(fn)
		if !isFunc(v) {
			return nil, fmt.Errorf("%w for key %q: %v", ErrInvalidFunc, k, v)
		}

		fns = append(fns, function{k, v})
	}

	return fns, nil
}

func isFunc(v reflect.Value) bool {
	return v.Kind() == reflect.Func
}

// FuncResult is a resulting of function execution.
type FuncResult struct {
	id   interface{}
	vals Values
	err  error
}

// Index returns the function index.
// The ok assu,es false if the function is not an item of the Funcs type.
func (f FuncResult) Index() (i int, ok bool) {
	i, ok = f.id.(int)
	return
}

// Name returns the function name.
// The ok assumes false if the function is not a value of the FuncMap type.
func (f FuncResult) Name() (name string, ok bool) {
	name, ok = f.id.(string)
	return
}

// Returned returns the values and error of a function execution.
func (f FuncResult) Returned() (Values, error) {
	return f.vals, f.err
}

// Waterfall executes all the functions in sequence, each function return value is passed to the next function.
// If the last or only return value of a function is of type error, this value is not passed to the next function.
//
// If the first parameter of any function is of type context.Context, the ctx passed to Waterfall will be used
// as the first argument of such functions, unless a previous function returns a context as its first value,
// in such cases the returned context will be used instead.
//
// The returned channel will receive either the last function return values or an error, if any.
// The execution flow stops in case a function last or only return value evaluates to a non-nil error.
//
// options is an optional list of Option for configuring Waterfall execution. Waterfall will use the given
// WithArgs values as the arguments of the first function.
func Waterfall(ctx context.Context, f Funcs, options ...Option) (<-chan FuncResult, error) {
	fns, err := f.funcs()
	if err != nil {
		return nil, err
	}

	opts := newOptions(options...)

	fnc := make(chan FuncResult, 1)

	go func() {
		defer close(fnc)

		var (
			fn  function
			err error
		)

		args := paramsValues(opts.args)

		for _, fn = range fns {
			select {
			case <-ctx.Done():
				fnc <- FuncResult{err: internal.NewContextError(ctx)}
				return
			default:
			}

			if args, err = execFunc(ctx, fn.Value, args); err != nil {
				break
			}
		}

		sendResult(fnc, fn.ID, args, err)
	}()

	return fnc, nil
}

// goroutines counts the number of goroutines ever created; for testing.
var goroutinesCount int32

// Concurrent executes all the functions concurrently. For each function a FuncResult message is sent through the channel
// with either the function return values or error, if any.
//
// For Concurrent to identify an error, the last or only return value of a function must evaluates to a non-nil error.
//
// In case the first parameter of any function is of type context.Context, the ctx passed to Concurrent will be used
// as the first argument of such functions, unless the first params item is a context.Context, in which case it takes
// precedence over the ctx and will be used instead.
//
// options is an optional list of Options for configuring Concurrent execution. The following options can be used with Concurrent:
//   - WithArgs values will be passed to all Concurrent functions.
//   - WithGoroutineNum value will be used to limit the number of goroutines created by Concurrent.
//
// The default number of goroutines to be created by Concurrent is number of functions to be executed.
func Concurrent(ctx context.Context, f FuncMap, options ...Option) (<-chan FuncResult, error) {
	fns, err := f.funcs()
	if err != nil {
		return nil, err
	}

	opts := newOptions(options...)

	fnc := make(chan FuncResult, len(fns))

	numGoroutines := goroutines(opts, len(fns))

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	args := paramsValues(opts.args)

	ctxErrc, notifyErrOnce := ctxErr()

	exec := func(in <-chan function) {
		defer wg.Done()

		atomic.AddInt32(&goroutinesCount, +1)

		for fn := range in {
			select {
			case <-ctx.Done():
				notifyErrOnce(ctx)
				return
			default:
			}

			res, err := execFunc(ctx, fn.Value, args)
			sendResult(fnc, fn.ID, res, err)
		}
	}

	in := genFuncs(fns...)
	for i := 0; i < numGoroutines; i++ {
		go exec(in)
	}

	go func() {
		wg.Wait()

		select {
		case err := <-ctxErrc:
			fnc <- FuncResult{err: err}
		default:
		}

		close(fnc)
	}()

	return fnc, nil
}

// Auto finds the best order to execute the functions based on their dependency relationship.
// If the last or only return value of a function is of type error, this value will not be passed to the dependents functions.
//
// In case the last or only return value of a function evaluates to a non-nil error, that error will be sent through the FuncResult
// channel and the functions that depend on the failed one will be not executed, if any.
//
// There must be no cyclic dependency relationship between the functions. In case of cyclic dependency relationship an
// ErrDependencyGraph is returned.
//
// params is an optional list of parameters to be passed to the functions without dependency.
//
// In case the first parameter of any function is of type context.Context, the ctx passed to Auto will be used
// as the first argument of such function, but the ctx takes the lowest precedence in the following cases:
//
// 1) For non-dependent functions: the first params item is a context.Context, in such cases that context will be used instead.
//
// 2) For dependent functions: the first function it depends on returns a context.Context as its first return value,
// in such cases the returned context will be used instead.
//
// options is an optional list of Options for configuring Auto execution. The following options can be used with Auto:
//   - WithArgs values will be passed to the functions without dependencies.
//   - WithGoroutineNum value will be used to limit the number of goroutines created by Auto.
//
// The default number of goroutines to be created by Auto is number of functions to be executed.
func Auto(ctx context.Context, fm FuncMap, options ...Option) (<-chan FuncResult, error) {
	opts := newOptions(options...)

	graph, err := funcsGraph(fm, opts.args)
	if err != nil {
		return nil, err
	}

	nodec, err := genSortedNodes(graph)
	if err != nil {
		return nil, err
	}

	fnc := make(chan FuncResult, len(fm))

	numGoroutines := goroutines(opts, len(fm))

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	ctxErrc, notifyErrOnce := ctxErr()

	exec := func(nodec <-chan *node) {
		defer wg.Done()

		atomic.AddInt32(&goroutinesCount, +1)

		for node := range nodec {
			select {
			case <-ctx.Done():
				notifyErrOnce(ctx)
				return
			default:
			}

			execNodeFunc(ctx, node, fnc, notifyErrOnce)
		}
	}

	for i := 0; i < numGoroutines; i++ {
		go exec(nodec)
	}

	go func() {
		wg.Wait()

		select {
		case err := <-ctxErrc:
			fnc <- FuncResult{err: err}
		default:
		}

		close(fnc)
	}()

	return fnc, nil
}

// DependentFunc is the representation of a dependent function used to express a relationship between functions.
type DependentFunc struct {
	deps Deps
	fn   interface{}
}

// Deps is a list of dependencies to be associated with functions.
type Deps []string

// To associates fn with the dependencies d, returning a DependentFunc.
func (d Deps) To(fn interface{}) DependentFunc {
	deps := make(Deps, len(d))
	copy(deps, d)

	return DependentFunc{
		deps: deps,
		fn:   fn,
	}
}

// DependsOn creates a list of dependencies to be associated with functions.
// Its arguments should be existent functions names within a depenency context.
func DependsOn(fns ...string) Deps {
	return fns
}

type node struct {
	name      string
	fn        reflect.Value
	outc      chan FuncResult
	edgesFrom []*node
	indegree  int
	outdegree int
}

type edges map[*node]map[*node]struct{}

func (e edges) from(n *node) (to map[*node]struct{}) {
	to = e[n]
	return
}

func (e edges) addAt(from, to *node, at int) {
	from.outdegree++
	to.indegree++

	to.edgesFrom[at] = from

	m, ok := e[from]
	if !ok {
		m = make(map[*node]struct{})
		e[from] = m
	}
	m[to] = struct{}{}
}

func (e edges) remove(from, to *node) {
	to.indegree--

	delete(e[from], to)
}

type graph struct {
	nodes map[string]*node
	edges edges
}

func (g *graph) nodeByName(name string) *node {
	n, ok := g.nodes[name]
	if !ok {
		n = &node{
			name: name,
			outc: make(chan FuncResult),
		}
		g.nodes[name] = n
	}
	return n
}

func (g *graph) addEdges(to *node, deps Deps) {
	to.edgesFrom = make([]*node, len(deps))
	for i, dep := range deps {
		from := g.nodeByName(dep)
		g.edges.addAt(from, to, i)
	}
}

// funcsGraph returns a graph from the dependency relationship describte in FuncMap.
func funcsGraph(fns FuncMap, params []interface{}) (*graph, error) {
	invalidFuncFmt := "%w for key %q: %v"

	grph := &graph{
		nodes: make(map[string]*node),
		edges: make(edges),
	}

	args := FuncResult{
		vals: paramsValues(params),
	}

	for name, fn := range fns {
		nde := grph.nodeByName(name)

		switch v := fn.(type) {
		case DependentFunc:
			nde.fn = reflect.ValueOf(v.fn)
			if !isFunc(nde.fn) {
				return nil, fmt.Errorf(invalidFuncFmt, ErrInvalidFunc, nde.name, v.fn)
			}
			grph.addEdges(nde, v.deps)
		default:
			nde.fn = reflect.ValueOf(fn)
			if !isFunc(nde.fn) {
				return nil, fmt.Errorf(invalidFuncFmt, ErrInvalidFunc, nde.name, fn)
			}

			d := &node{outc: make(chan FuncResult, 1)}
			d.outc <- args

			nde.edgesFrom = []*node{d}
		}
	}

	for _, node := range grph.nodes {
		if dependentsNum := node.outdegree; dependentsNum > 0 {
			node.outc = make(chan FuncResult, dependentsNum)
		}
	}

	return grph, nil
}

// genSortedNodes returns a channel of sorted nodes from the given graph g.
// It uses the Kahn's topological sort algorithm. See https://en.wikipedia.org/wiki/Topological_sorting#Kahn's_algorithm.
// it returns a ErrDependencyGraph in case of a cyclic graph.
func genSortedNodes(g *graph) (<-chan *node, error) {
	sortc := make(chan *node, len(g.nodes))
	defer close(sortc)

	var noIndegrees []*node
	for _, n := range g.nodes {
		if n.indegree == 0 {
			noIndegrees = append(noIndegrees, n)
		}
	}

	for len(noIndegrees) > 0 {
		var n *node
		n, noIndegrees = noIndegrees[0], noIndegrees[1:]

		sortc <- n

		for to := range g.edges.from(n) {
			g.edges.remove(n, to)

			if to.indegree == 0 {
				noIndegrees = append(noIndegrees, to)
			}
		}
	}

	if len(g.nodes) != len(sortc) {
		return nil, ErrDependencyGraph
	}

	return sortc, nil
}

func goroutines(opts *Options, defaultNum int) int {
	if n := opts.goroutineNum; n > 0 {
		return n
	}

	return defaultNum
}

func genFuncs(fns ...function) <-chan function {
	c := make(chan function, len(fns))
	defer close(c)

	for _, fn := range fns {
		c <- fn
	}

	return c
}

func paramsValues(params []interface{}) []reflect.Value {
	values := make([]reflect.Value, len(params))
	for i, param := range params {
		values[i] = reflect.ValueOf(param)
	}
	return values
}

func execFunc(ctx context.Context, fn reflect.Value, args []reflect.Value) (res []reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v", ErrPanicked, r)
		}
	}()

	if expectsContext(fn) {
		args = argsWithContext(ctx, args)
	}

	res = fn.Call(args)

	if returnsErr(fn) {
		l := len(res) - 1
		if verr := res[l]; !verr.IsNil() {
			return nil, fmt.Errorf("[async] function returned: %w", verr.Interface().(error))
		}

		res = res[:l]
	}

	return res, nil
}

func execNodeFunc(ctx context.Context, n *node, fnc chan<- FuncResult, notifyErrOnce func(context.Context)) {
	defer close(n.outc)

	args, ok := depsResult(ctx, n.edgesFrom, notifyErrOnce)
	if !ok {
		return
	}

	res, err := execFunc(ctx, n.fn, args)
	if err != nil {
		sendResult(fnc, n.name, res, err)
		return
	}

	if n.outdegree == 0 {
		sendResult(fnc, n.name, res, nil)
	} else {
		propagateResult(n, res)
	}
}

var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()

func expectsContext(fn reflect.Value) bool {
	if fnType := fn.Type(); fnType.NumIn() > 0 {
		return fnType.In(0).Implements(contextType)
	}

	return false
}

func argsWithContext(ctx context.Context, args []reflect.Value) []reflect.Value {
	if len(args) == 0 || !args[0].Type().Implements(contextType) {
		args = append([]reflect.Value{reflect.ValueOf(ctx)}, args...)
	}
	return args
}

func depsResult(ctx context.Context, deps []*node, notifyErrOnce func(context.Context)) (res []reflect.Value, ok bool) {
	for _, dep := range deps {
		var r FuncResult

		select {
		case <-ctx.Done():
			notifyErrOnce(ctx)
			ok = false
			return res, ok
		case r, ok = <-dep.outc:
			if !ok {
				return res, ok
			}
			res = append(res, r.vals...)
		}
	}

	return res, ok
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func returnsErr(fn reflect.Value) bool {
	fnType := fn.Type()

	if fnType.NumOut() == 0 {
		return false
	}

	return fnType.Out(fnType.NumOut() - 1).Implements(errorType)
}

func sendResult(fnc chan<- FuncResult, id interface{}, vals Values, err error) {
	fnc <- FuncResult{
		id:   id,
		vals: vals,
		err:  err,
	}
}

func propagateResult(n *node, res Values) {
	for i := 0; i < n.outdegree; i++ {
		sendResult(n.outc, n.name, res, nil)
	}
}
