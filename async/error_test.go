package async

import (
	"context"
	"errors"
	"testing"

	"github.com/rdleal/go-async/async/internal"
)

func TestError(t *testing.T) {
	dummyErr := err("dummy error")

	if got, want := dummyErr.Error(), "[async] dummy error"; got != want {
		t.Errorf("Error() = got error message: %q; want %q.", got, want)
	}
}

func TestIsContextError(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "True",
			err:  internal.NewContextError(context.Background()),
			want: true,
		},
		{
			name: "False",
			err:  errors.New("Custom error"),
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsContextError(tc.err); got != tc.want {
				t.Errorf("Got IsContextError(%v) = %t; want %t.", tc.err, got, tc.want)
			}
		})
	}
}
