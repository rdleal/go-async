package async

import (
	"fmt"
	"reflect"
	"strings"
)

// Values represents a list of a function return values.
type Values []reflect.Value

// BoolAt returns i-th v's underlying value.
// The ok return value assumes false if i-th v's kind is not Bool.
// It panics if i is out of v's range.
func (v Values) BoolAt(i int) (t, ok bool) {
	val := v[i]
	if ok = val.Kind() == reflect.Bool; !ok {
		return
	}

	t = val.Bool()

	return
}

// BytesAt returns i-th v's underlying value.
// The ok return value assumes false if i-th v's underlying value is not a slice of bytes.
// It panics if i is out of v's range.
func (v Values) BytesAt(i int) (b []byte, ok bool) {
	val := v[i]
	if ok = val.Kind() == reflect.Slice && val.Type().Elem().Kind() == reflect.Uint8; !ok {
		return
	}

	b = val.Bytes()

	return
}

// ComplexAt returns i-th v's underlying value, as a complex128.
// The ok return value assumes false if i-th v's Kind is not Complex64 or Complex128.
// It panics if i is out of v's range.
func (v Values) ComplexAt(i int) (c complex128, ok bool) {
	switch val := v[i]; val.Kind() {
	case reflect.Complex64, reflect.Complex128:
		c, ok = val.Complex(), true
	}

	return
}

// FloatAt returns i-th v's underlying value, as a float64.
// The ok return value assumes false if i-th v's Kind is not Float32 or Float64.
// It panics if i is out of v's range.
func (v Values) FloatAt(i int) (f float64, ok bool) {
	switch val := v[i]; val.Kind() {
	case reflect.Float32, reflect.Float64:
		f, ok = val.Float(), true
	}

	return
}

// IntAt returns i-th v's underlying value, as an int64.
// The ok return value assumes false if i-th v's Kind is not Int, Int8, Int16, Int32, or Int64.
// It panics if i is out of v's range.
func (v Values) IntAt(i int) (n int64, ok bool) {
	switch val := v[i]; val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, ok = val.Int(), true
	}

	return
}

// InterfaceAt returns i-th v's current value as an interface{}.
// The ok return value assumes false if either i=th v is not valid or the interface cannot be used without panicking.
// It panics if i is out of v's range.
func (v Values) InterfaceAt(i int) (r interface{}, ok bool) {
	val := v[i]
	if ok = val.IsValid(); !ok {
		return
	}

	if ok = val.CanInterface(); !ok {
		return
	}

	r = val.Interface()

	return
}

// PointerAt returns i-th v's value as a uintptr.
// The ok return value assumes false if i-th v's Kind is not Chan, Func, Map, Ptr, Slice, or UnsafePointer.
// It panics if i is out of v's range.
func (v Values) PointerAt(i int) (p uintptr, ok bool) {
	switch val := v[i]; val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		p, ok = val.Pointer(), true
	}

	return
}

// StringAt returns the string i-th v's underlying value, as a string.
// The ok return value assumes false if i-th v's Kind is not String.
// It panics if i is out of v's range.
func (v Values) StringAt(i int) (s string, ok bool) {
	val := v[i]
	if ok = val.Kind() == reflect.String; !ok {
		return
	}

	s = val.String()

	return
}

// UintAt returns i-th v's underlying value, as a uint64.
// The ok return value assumes false if v's Kind is not Uint, Uintptr, Uint8, Uint16, Uint32, or Uint64.
// It panics if i is out of v's range.
func (v Values) UintAt(i int) (u uint64, ok bool) {
	switch val := v[i]; val.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, ok = val.Uint(), true
	}

	return
}

// UnsafeAddrAt returns a pointer to i-th v's data.
// The ok return value assumes false if i-th v is not addressable.
// It panics if i is out of v's range.
func (v Values) UnsafeAddrAt(i int) (p uintptr, ok bool) {
	val := v[i]
	if ok = val.CanAddr(); !ok {
		return
	}

	p = val.UnsafeAddr()

	return
}

const emptyValuesStr = "[]"
const sep = ' '

// String returns the v's underlying values, as string.
func (v Values) String() string {
	switch len(v) {
	case 0:
		return emptyValuesStr
	case 1:
		return fmt.Sprintf("[%v]", v[0])
	}

	var b strings.Builder
	b.WriteByte('[')
	fmt.Fprintf(&b, "%v", v[0])
	for _, val := range v[1:] {
		b.WriteByte(sep)
		fmt.Fprintf(&b, "%v", val)
	}
	b.WriteByte(']')
	return b.String()
}
