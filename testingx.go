// Package testingx
package testingx

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var errMatcher func(got, want any) bool = func(got, want any) bool {
	return errors.Is(got.(error), want.(error))
}

func SetDefaultErrMatcher(f func(got, want any) bool) {
	errMatcher = f
}

func TypeMismatch(t testing.TB, got, want any) {
	t.Helper()
	t.Errorf("got type %T, want %T", got, want)
}

func ErrUnexpected(t testing.TB, got error) {
	t.Helper()
	t.Errorf("unexpected error: %v", got)
}

func ErrExpected(t testing.TB, want any) {
	t.Helper()
	t.Errorf("expected error %v, got nil", want)
}

func ErrMismatch(t testing.TB, got error, want any) {
	t.Helper()
	t.Errorf("got error %v, want error %v", got, want)
}

func Mismatch(t testing.TB, diff string) {
	t.Helper()
	t.Errorf("mismatch (-want +got):\n%s", diff)
}

func Check(t testing.TB, got, want any, opts ...cmp.Option) {
	t.Helper()
	if reflect.TypeOf(got) != reflect.TypeOf(want) {
		TypeMismatch(t, got, want)
		return
	}
	if diff := cmp.Diff(want, got, opts...); diff != "" {
		Mismatch(t, diff)
	}
}

func CheckErr(t testing.TB, got, want any, gotErr error, wantErr any, opts ...cmp.Option) {
	t.Helper()
	switch {
	case gotErr != nil && wantErr != nil:
		if !errMatcher(gotErr, wantErr) {
			ErrMismatch(t, gotErr, wantErr)
		}
	case gotErr != nil && wantErr == nil:
		ErrUnexpected(t, gotErr)
	case gotErr == nil && wantErr != nil:
		ErrExpected(t, wantErr)
	case gotErr == nil && wantErr == nil:
		Check(t, got, want, opts...)
	}
}

func CheckPanic(t testing.TB, fn func(), expectedPanic any) (panicked bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if r != expectedPanic {
				panic(r)
			}
			panicked = true
		}
	}()
	fn()
	return
}
