# go-async

[![PkgGoDev](https://pkg.go.dev/badge/github.com/rdleal/go-async/async)](https://pkg.go.dev/github.com/rdleal/go-async/async)
[![Go Report Card](https://goreportcard.com/badge/github.com/rdleal/go-async)](https://goreportcard.com/report/github.com/rdleal/go-async)

go-async provides utility functions for controlling asynchronous flow.

## Installing

Install the package:
```
go get github.com/rdleal/go-async
``` 

## Usage

Import the package as:
```
import "github.com/rdleal/go-async/async"
```

### Concurrent

Concurrent function is mainly useful when executing unordered functions.

```go
package main

import (
        "context"
        "fmt"
        "log"

        "github.com/rdleal/go-async/async"
)

func main() {
	ch := make(chan int)
	fnc, err := async.Concurrent(context.Background(), async.FuncMap{
		"gen": func(out chan<- int) {
			defer close(out)

			for _, n := range []int{2, 3} {
				out <- n
			}
		},
		"sq": func(in <-chan int) []int {
			var squared []int
			for n := range in {
				squared = append(squared, n*n)
			}
			return squared
		},
	}, async.WithArgs(ch))
	if err != nil {
		log.Fatal(err)
	}

	for fn := range fnc {
		res, err := fn.Returned()
		if err != nil {
			log.Fatal(err)
		}

		name, ok := fn.Name()
		if !ok {
			log.Fatal("Function without name")
		}

		if name == "sq" {
			fmt.Println(res[0]) // [4 9]
		}
	}
}

```

### Waterfall

Waterfall is useful for sequencial execution of functions:

```go
import (
        "context"
        "fmt"
        "log"
	"math/big"

        "github.com/rdleal/go-async/async"
)

func main() {
	fibs := func(n int) chan int {
		fibc := make(chan int, n)
		go func() {
			defer close(fibc)

			for i, j := 0, 1; i < n; i, j = i+j, i {
				fibc <- i
			}
		}()
		return fibc
	}

	primes := func(in <-chan int) []int {
		var primes []int
		for n := range in {
			if big.NewInt(int64(n)).ProbablyPrime(0) {
				primes = append(primes, n)
			}
		}
		return primes
	}

	fnc, err := async.Waterfall(context.Background(), async.Funcs{
		fibs,
		primes,
	}, async.WithArgs(1000))
	if err != nil {
		log.Fatal(err)
	}

	fn := <-fnc

	res, err := fn.Returned()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res[0]) // [2 3 5 13 89 233]

}
```

### Auto

Auto finds the best order of function execution based on their dependency relationship:

```go
package main

import (
        "context"
        "encoding/json"
        "fmt"
        "log"

        "github.com/rdleal/go-async/async"
)

func main() { 
	type msg struct {
		Greet, Name string
	}

	const jsonStr = `{"greet": "hello", "name" : "gopher"}`
	fnc, err := async.Auto(context.Background(), async.FuncMap{
		"reader":  strings.NewReader,
		"decoder": async.DependsOn("reader").To(json.NewDecoder),
		"parser": async.DependsOn("decoder").To(func(d *json.Decoder) (msg, error) {
			var m msg
			err := d.Decode(&m)
			return m, err
		}),
		"greet": async.DependsOn("parser").To(func(m msg) string {
			return strings.Title(m.Greet)
		}),
		"name": async.DependsOn("parser").To(func(m msg) string {
			return strings.Title(m.Name)
		}),
		"greeter": async.DependsOn("greet", "name").To(func(greet, name string) string {
			return fmt.Sprintf("%s, %s!", greet, name)
		}),
	}, async.WithArgs(jsonStr))
	if err != nil {
		log.Fatal(err)
	}

	for fn := range fnc {
		res, err := fn.Returned()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res[0]) // Hello, Gopher!
	}
} 
``` 

## License

Distributed under MIT License. See [LICENSE](LICENSE) file for more details.
