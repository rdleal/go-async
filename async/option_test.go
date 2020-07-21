package async

import (
	"reflect"
	"testing"
)

func TestOptions(t *testing.T) {
	testCases := []struct {
		name   string
		option Option
		want   *Options
	}{
		{
			name:   "WithArgs",
			option: WithArgs("Hello", "Gopher", 1),
			want: &Options{
				args: []interface{}{"Hello", "Gopher", 1},
			},
		},
		{
			name:   "WithGoroutineNum",
			option: WithGoroutineNum(3),
			want: &Options{
				goroutineNum: 3,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := new(Options)
			tc.option(got)

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Got Options %v; want %v.", got, tc.want)
			}
		})
	}
}
