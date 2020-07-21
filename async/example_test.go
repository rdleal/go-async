package async_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/rdleal/go-async/async"
)

func ExampleWaterfall() {
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
		strconv.Atoi,
		fibs,
		primes,
	}, async.WithArgs("1000"))
	if err != nil {
		log.Fatal(err)
	}

	fn := <-fnc

	res, err := fn.Returned()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res[0])
	// Output:
	// [2 3 5 13 89 233]
}

func ExampleConcurrent() {
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
			fmt.Println(res[0])
		}
		// Output:
		// [4 9]
	}
}

func ExampleConcurrent_race() {
	fnc, err := async.Concurrent(context.Background(), async.FuncMap{
		"Java":   func() { time.Sleep(time.Millisecond * 100) },
		"Golang": func() { time.Sleep(time.Millisecond * 10) },
	})
	if err != nil {
		log.Fatal(err)
	}

	fn := <-fnc

	name, ok := fn.Name()
	if !ok {
		log.Fatal("Function without name")
	}

	fmt.Printf("Winner %q", name)
	// Output:
	// Winner "Golang"
}

func ExampleAuto() {
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
		fmt.Println(res[0])
		// Output:
		// Hello, Gopher!
	}
}
