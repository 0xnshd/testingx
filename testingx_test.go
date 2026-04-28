package testingx_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/0xnshd/testingx"
	"github.com/google/go-cmp/cmp"
)

const (
	expectedFailureMessage    = "expected failure did not occur"
	wrongFailureMessageFormat = "wrong failure message: %s, expected: %s"
)

type spyT struct {
	testing.TB
	failed  bool
	message string
}

func (s *spyT) Errorf(format string, args ...any) {
	s.failed = true
	s.message = fmt.Sprintf(format, args...)
}

func (s *spyT) Helper() {}

func spyRun(fn func()) {
	var wg sync.WaitGroup
	wg.Go(fn)
	wg.Wait()
}

func Test_CheckWErr(t *testing.T) {
	errA := errors.New("a")
	errB := errors.New("b")

	tests := []struct {
		name        string
		got         any
		want        any
		err         error
		wantErr     error
		wantMessage string
	}{
		{
			name:        "ErrUnexpected",
			got:         nil,
			want:        nil,
			err:         errA,
			wantErr:     nil,
			wantMessage: fmt.Sprintf("unexpected error: %v", errA),
		},
		{
			name:        "ErrExpected",
			got:         nil,
			want:        nil,
			err:         nil,
			wantErr:     errA,
			wantMessage: fmt.Sprintf("expected error %v, got nil", errA),
		},
		{
			name:        "ErrMismatch",
			got:         nil,
			want:        nil,
			err:         errA,
			wantErr:     errB,
			wantMessage: fmt.Sprintf("got error %v, want error %v", errA, errB),
		},
		{
			name:        "Check_TypeMismatch",
			got:         1,
			want:        "1",
			err:         nil,
			wantErr:     nil,
			wantMessage: fmt.Sprintf("got type %T, want %T", 1, "1"),
		},
		{
			name:        "Check_Mismatch",
			got:         "ab",
			want:        "abc",
			err:         nil,
			wantErr:     nil,
			wantMessage: fmt.Sprintf("mismatch (-want +got):\n%s", cmp.Diff("abc", "ab")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spy := &spyT{TB: t}
			spyRun(func() {
				testingx.CheckWErr(spy, tc.got, tc.want, tc.err, tc.wantErr)
			})

			if !spy.failed {
				t.Error(expectedFailureMessage)
			}
			if spy.message != tc.wantMessage {
				t.Errorf(wrongFailureMessageFormat, spy.message, tc.wantMessage)
			}
		})
	}
}

func Test_CheckPanic(t *testing.T) {
	tests := []struct {
		name          string
		fn            func()
		expectedPanic any
		wantReturn    bool
		wantRecover   string
	}{
		{
			name:          "Expected panic",
			fn:            func() { panic("abc") },
			expectedPanic: "abc",
			wantReturn:    true,
			wantRecover:   "",
		},
		{
			name:          "Unexpected panic",
			fn:            func() { panic("abc") },
			expectedPanic: "ab",
			wantReturn:    false,
			wantRecover:   "abc",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			done := make(chan string, 1)
			var gotReturn bool
			go func() {
				defer func() {
					if r := recover(); r != nil {
						done <- fmt.Sprintf("%v", r)
					} else {
						done <- ""
					}
				}()
				gotReturn = testingx.CheckPanic(t, tc.fn, tc.expectedPanic)
			}()

			gotRecover := <-done

			if gotReturn != tc.wantReturn {
				t.Errorf("want return %v, got %v", tc.wantRecover, gotReturn)
			}

			if gotRecover != tc.wantRecover {
				t.Errorf("expected re-panic with %s, got %s", tc.wantRecover, gotRecover)
			}
		})
	}
}

func Test_SetDefaultErrMatcher(t *testing.T) {
	errA := errors.New("a")
	errB := errors.New("b")

	defaultMatcher := func(got, want error) bool {
		return errors.Is(got, want)
	}

	tests := []struct {
		name        string
		matcher     func(error, error) bool
		wantFailed  bool
		wantMessage string
	}{
		{
			name:        "Rejects",
			matcher:     func(got, want error) bool { return false },
			wantFailed:  true,
			wantMessage: fmt.Sprintf("got error %v, want error %v", errA, errB),
		},
		{
			name:        "Accepts",
			matcher:     func(got, want error) bool { return true },
			wantFailed:  false,
			wantMessage: "", // no message check for this case
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() { testingx.SetDefaultErrMatcher(defaultMatcher) })
			testingx.SetDefaultErrMatcher(tc.matcher)

			spy := &spyT{TB: t}
			spyRun(func() {
				testingx.CheckWErr(spy, nil, nil, errA, errB)
			})

			if spy.failed != tc.wantFailed {
				t.Errorf("expected failed: %v, got %v", tc.wantFailed, spy.failed)
			}
			if spy.message != tc.wantMessage {
				t.Errorf(wrongFailureMessageFormat, spy.message, tc.wantMessage)
			}
		})
	}
}
