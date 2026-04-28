# testingx
A thin, composable testing utility for Go that wraps `testing.TB` with ergonomic helpers for value comparison, error assertion, and panic checking.

## Install
```bash
go get github.com/0xnshd/testingx
```

## Usage

All functions accept `testing.TB` as the first argument, so they work with `*testing.T`, `*testing.B`, and subtests without wrapping.

---

### Check — value comparison

Uses `cmp.Diff` under the hood. Arguments are `got, want`. Also validates that both share the same type before diffing.

```go
testingx.Check(t, got, want)
// with cmp options
testingx.Check(t, got, want, cmpopts.IgnoreUnexported(MyStruct{}))
```

---

### CheckErr — error comparison

Compares two errors using the configured matcher (defaults to `errors.Is`).

```go
testingx.CheckErr(t, gotErr, wantErr)
```

---

### CheckWErr — value + error in one call

Handles all 4 combinations of `gotErr` and `wantErr`.

```go
got, gotErr := DoSomething()
testingx.CheckWErr(t, got, want, gotErr, wantErr)
```

| gotErr | wantErr | behavior |
|--------|---------|----------|
| non-nil | non-nil | runs `CheckErr` |
| non-nil | nil | unexpected error |
| nil | non-nil | expected error, got nil |
| nil | nil | runs `Check` |

#### Custom error matcher

By default uses `errors.Is`. Override globally for custom error types:

```go
testingx.SetDefaultErrMatcher(func(got, want error) bool {
    return serror.OfTrait(got, want.(serror.Trait))
})
```

---

### CheckPanic — panic assertion

Runs `fn`, catches the expected panic, re-panics on anything unexpected. Returns `true` if the expected panic occurred, `false` if `fn` ran cleanly.

```go
if testingx.CheckPanic(t, func() {
    result = DoSomething(arg)
}, ErrExpectedPanic) {
    return // or continue in table test
}
```

---

### Low-level helpers

For cases where the higher-level helpers create friction:

```go
testingx.TypeMismatch(t, got, (*MyStruct)(nil))  // got type *Foo, want *MyStruct
testingx.ErrUnexpected(t, gotErr)                // unexpected error: ...
testingx.ErrExpected(t, wantErr)                 // expected error ..., got nil
testingx.ErrMismatch(t, gotErr, wantErr)         // got error ..., want error ...
testingx.Mismatch(t, diff)                       // mismatch (-want +got): ...
```

Useful when asserting concrete types before further checks:

```go
got, ok := gotE.(*ErrorRecord)
if !ok {
    testingx.TypeMismatch(t, gotE, (*ErrorRecord)(nil))
    return
}
```

---

## Full example — table test

```go
func TestNew(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        want      *Result
        wantErr   error
        wantPanic any
    }{
        // ...
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var got *Result
            var err error
            if testingx.CheckPanic(t, func() {
                got, err = New(tt.input)
            }, tt.wantPanic) {
                return
            }
            testingx.CheckWErr(t, got, tt.want, err, tt.wantErr)
        })
    }
}
```
