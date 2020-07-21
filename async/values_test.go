package async

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestValues_BoolAt(t *testing.T) {
	testCases := []struct {
		desc   string
		idx    int
		vals   Values
		want   bool
		wantOK bool
	}{
		{
			desc: "TrueFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(true),
			},
			want:   true,
			wantOK: true,
		},
		{
			desc: "FalseFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(false),
			},
			want:   false,
			wantOK: true,
		},
		{
			desc: "BoolNotFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(42),
			},
			want:   false,
			wantOK: false,
		},
		{
			desc: "InvalidType",
			idx:  0,
			vals: Values{
				reflect.Value{},
			},
			want:   false,
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, ok := tc.vals.BoolAt(tc.idx)
			if ok != tc.wantOK {
				t.Errorf("Got ok value: %t; want %t.", ok, tc.wantOK)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Got value: %v; want %v.", got, tc.want)
			}
		})
	}
}

func TestValues_BytesAt(t *testing.T) {
	testCases := []struct {
		desc   string
		idx    int
		vals   Values
		want   []byte
		wantOK bool
	}{
		{
			desc: "BytesFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf([]byte("Golang")),
			},
			want:   []byte("Golang"),
			wantOK: true,
		},
		{
			desc: "BytesNotFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(42),
			},
			want:   nil,
			wantOK: false,
		},
		{
			desc: "InvalidType",
			idx:  0,
			vals: Values{
				reflect.Value{},
			},
			want:   nil,
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, ok := tc.vals.BytesAt(tc.idx)
			if ok != tc.wantOK {
				t.Errorf("Got ok value: %t; want %t.", ok, tc.wantOK)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Got value: %v; want %v.", got, tc.want)
			}
		})
	}
}

func TestValues_ComplexAt(t *testing.T) {
	testCases := []struct {
		desc   string
		idx    int
		vals   Values
		want   complex128
		wantOK bool
	}{
		{
			desc: "Complex64Found",
			idx:  0,
			vals: Values{
				reflect.ValueOf(complex(float32(1), float32(4))),
			},
			want:   1 + 4i,
			wantOK: true,
		},
		{
			desc: "Complex128Found",
			idx:  0,
			vals: Values{
				reflect.ValueOf(1 + 4i),
			},
			want:   1 + 4i,
			wantOK: true,
		},
		{
			desc: "ComplexNotFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf("Golang"),
			},
			want:   0,
			wantOK: false,
		},
		{
			desc: "InvalidType",
			idx:  0,
			vals: Values{
				reflect.Value{},
			},
			want:   0,
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, ok := tc.vals.ComplexAt(tc.idx)
			if ok != tc.wantOK {
				t.Errorf("Got ok value: %t; want %t.", ok, tc.wantOK)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Got value: %v; want %v.", got, tc.want)
			}
		})
	}
}

func TestValues_FloatAt(t *testing.T) {
	testCases := []struct {
		desc   string
		idx    int
		vals   Values
		want   float64
		wantOK bool
	}{
		{
			desc: "Float32Found",
			idx:  0,
			vals: Values{
				reflect.ValueOf(float32(42)),
			},
			want:   42,
			wantOK: true,
		},
		{
			desc: "Float64Found",
			idx:  0,
			vals: Values{
				reflect.ValueOf(64.0),
			},
			want:   64,
			wantOK: true,
		},
		{
			desc: "FloatNotFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf("Golang"),
			},
			want:   0,
			wantOK: false,
		},
		{
			desc: "InvalidType",
			idx:  0,
			vals: Values{
				reflect.Value{},
			},
			want:   0,
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, ok := tc.vals.FloatAt(tc.idx)
			if ok != tc.wantOK {
				t.Errorf("Got ok value: %t; want %t.", ok, tc.wantOK)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Got value: %v; want %v.", got, tc.want)
			}
		})
	}
}

func TestValues_IntAt(t *testing.T) {
	testCases := []struct {
		desc   string
		idx    int
		vals   Values
		want   int64
		wantOK bool
	}{
		{
			desc: "IntFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(42),
			},
			want:   42,
			wantOK: true,
		},
		{
			desc: "Int8Found",
			idx:  0,
			vals: Values{
				reflect.ValueOf(int8(4)),
			},
			want:   4,
			wantOK: true,
		},
		{
			desc: "Int16Found",
			idx:  0,
			vals: Values{
				reflect.ValueOf(int16(2)),
			},
			want:   2,
			wantOK: true,
		},
		{
			desc: "Int32Found",
			idx:  0,
			vals: Values{
				reflect.ValueOf(int32(6)),
			},
			want:   6,
			wantOK: true,
		},
		{
			desc: "Int64Found",
			idx:  0,
			vals: Values{
				reflect.ValueOf(int64(6)),
			},
			want:   6,
			wantOK: true,
		},
		{
			desc: "IntNotFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf("Golang"),
			},
			want:   0,
			wantOK: false,
		},
		{
			desc: "InvalidType",
			idx:  0,
			vals: Values{
				reflect.Value{},
			},
			want:   0,
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, ok := tc.vals.IntAt(tc.idx)
			if ok != tc.wantOK {
				t.Errorf("Got ok value: %t; want %t.", ok, tc.wantOK)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Got value: %v; want %v.", got, tc.want)
			}
		})
	}
}

func TestValues_InterfaceAt(t *testing.T) {
	testCases := []struct {
		desc   string
		idx    int
		vals   Values
		want   interface{}
		wantOK bool
	}{
		{
			desc: "InterfaceFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(struct{}{}),
			},
			want:   struct{}{},
			wantOK: true,
		},
		{
			desc: "InvalidType",
			idx:  0,
			vals: Values{
				reflect.Value{},
			},
			want:   nil,
			wantOK: false,
		},
		{
			desc: "InvalidInterfaceType",
			idx:  0,
			vals: Values{
				reflect.ValueOf(struct{ field string }{}).FieldByName("field"),
			},
			want:   nil,
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, ok := tc.vals.InterfaceAt(tc.idx)
			if ok != tc.wantOK {
				t.Errorf("Got ok value: %t; want %t.", ok, tc.wantOK)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Got value: %v; want %v.", got, tc.want)
			}
		})
	}
}

func TestValues_Pointer(t *testing.T) {
	var (
		empty struct{}
		slice = make([]int, 2)
		mp    = make(map[string]int)
		ch    = make(chan struct{})
		fn    = func() {}
	)

	testCases := []struct {
		desc   string
		idx    int
		vals   Values
		want   uintptr
		wantOK bool
	}{
		{
			desc: "PointerFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(&empty),
			},
			want:   reflect.ValueOf(&empty).Pointer(),
			wantOK: true,
		},
		{
			desc: "PointerOfSliceFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(slice),
			},
			want:   reflect.ValueOf(slice).Pointer(),
			wantOK: true,
		},
		{
			desc: "PointerOfMapFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(mp),
			},
			want:   reflect.ValueOf(mp).Pointer(),
			wantOK: true,
		},
		{
			desc: "PointerOfChannelFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(ch),
			},
			want:   reflect.ValueOf(ch).Pointer(),
			wantOK: true,
		},
		{
			desc: "PointerOfFunctionFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(fn),
			},
			want:   reflect.ValueOf(fn).Pointer(),
			wantOK: true,
		},
		{
			desc: "PointerOfUnsafePointerFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(unsafe.Pointer(&empty)),
			},
			want:   reflect.ValueOf(unsafe.Pointer(&empty)).Pointer(),
			wantOK: true,
		},
		{
			desc: "PointerNotFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf("Not a Pointer"),
			},
			want:   0,
			wantOK: false,
		},
		{
			desc: "InvalidType",
			idx:  0,
			vals: Values{
				reflect.Value{},
			},
			want:   0,
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, ok := tc.vals.PointerAt(tc.idx)
			if ok != tc.wantOK {
				t.Errorf("Got ok value: %t; want %t.", ok, tc.wantOK)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Got value: %v; want %v.", got, tc.want)
			}
		})
	}
}

func TestValues_StringAt(t *testing.T) {
	testCases := []struct {
		desc   string
		idx    int
		vals   Values
		want   string
		wantOK bool
	}{
		{
			desc: "StringFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf("Golang is awesome"),
			},
			want:   "Golang is awesome",
			wantOK: true,
		},
		{
			desc: "StringNotFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(42),
			},
			want:   "",
			wantOK: false,
		},
		{
			desc: "InvalidType",
			idx:  0,
			vals: Values{
				reflect.Value{},
			},
			want:   "",
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, ok := tc.vals.StringAt(tc.idx)
			if ok != tc.wantOK {
				t.Errorf("Got ok value: %t; want %t.", ok, tc.wantOK)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Got value: %v; want %v.", got, tc.want)
			}
		})
	}
}

func TestValues_UintAt(t *testing.T) {
	testCases := []struct {
		desc   string
		idx    int
		vals   Values
		want   uint64
		wantOK bool
	}{
		{
			desc: "UintFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(uint(42)),
			},
			want:   42,
			wantOK: true,
		},
		{
			desc: "Uint8Found",
			idx:  0,
			vals: Values{
				reflect.ValueOf(uint8(4)),
			},
			want:   4,
			wantOK: true,
		},
		{
			desc: "Uint16Found",
			idx:  0,
			vals: Values{
				reflect.ValueOf(uint16(2)),
			},
			want:   2,
			wantOK: true,
		},
		{
			desc: "Uint32Found",
			idx:  0,
			vals: Values{
				reflect.ValueOf(uint32(6)),
			},
			want:   6,
			wantOK: true,
		},
		{
			desc: "Uint64Found",
			idx:  0,
			vals: Values{
				reflect.ValueOf(uint64(6)),
			},
			want:   6,
			wantOK: true,
		},
		{
			desc: "UintNotFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf("Golang"),
			},
			want:   0,
			wantOK: false,
		},
		{
			desc: "InvalidType",
			idx:  0,
			vals: Values{
				reflect.Value{},
			},
			want:   0,
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, ok := tc.vals.UintAt(tc.idx)
			if ok != tc.wantOK {
				t.Errorf("Got ok value: %t; want %t.", ok, tc.wantOK)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Got value: %v; want %v.", got, tc.want)
			}
		})
	}
}

func TestValues_UnsafeAddrAt(t *testing.T) {
	var empty struct{}

	testCases := []struct {
		desc   string
		idx    int
		vals   Values
		want   uintptr
		wantOK bool
	}{
		{
			desc: "UnsafeAddrFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(&empty).Elem(),
			},
			want:   uintptr(unsafe.Pointer(&empty)),
			wantOK: true,
		},
		{
			desc: "UnsafeAddrNotFound",
			idx:  0,
			vals: Values{
				reflect.ValueOf(empty),
			},
			want:   0,
			wantOK: false,
		},
		{
			desc: "InvalidType",
			idx:  0,
			vals: Values{
				reflect.Value{},
			},
			want:   0,
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, ok := tc.vals.UnsafeAddrAt(tc.idx)
			if ok != tc.wantOK {
				t.Errorf("Got ok value: %t; want %t.", ok, tc.wantOK)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Got value: %v; want %v.", got, tc.want)
			}
		})
	}
}

func TestValues_String(t *testing.T) {
	testCases := []struct {
		name string
		vals Values
		want string
	}{
		{
			name: "WithValues",
			vals: Values{
				reflect.ValueOf("testing"),
				reflect.ValueOf(42),
				reflect.ValueOf(true),
			},
			want: `[testing 42 true]`,
		},
		{
			name: "OneValue",
			vals: Values{reflect.ValueOf("Golang")},
			want: `[Golang]`,
		},
		{
			name: "EmptyValues",
			vals: Values{},
			want: `[]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.vals.String()

			if got != tc.want {
				t.Errorf("Got string Values: %q; want %q.", got, tc.want)
			}

		})
	}
}
