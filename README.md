# testingx

A thin, composable testing utility for Go that wraps `*testing.T` with ergonomic helpers for value comparison, error assertion, and panic checking.

## Install

```bash
go get github.com/0xnshd/testingx
```

## Usage

### Setup

```go
func TestSomething(t *testing.T) {
    tx := testingx.New(t)
    // use tx instead of t
}
```

---

### Check — value comparison

Uses `cmp.Diff` under the hood. Arguments are `want, got` to match diff output convention.

```go
tx.Check(want, got)

// with cmp options
tx.Check(want, got, cmpopts.IgnoreUnexported(MyStruct{}))
```

---

### CheckErr — error + value in one call

Handles all 4 combinations of `gotErr` and `wantErr`.

```go
got, gotErr := DoSomething()
tx.CheckErr(got, want, gotErr, wantErr)
```

| gotErr | wantErr | behavior |
|--------|---------|----------|
| non-nil | non-nil | runs errMatcher |
| non-nil | nil | unexpected error |
| nil | non-nil | expected error, got nil |
| nil | nil | runs Check |

#### Custom error matcher

By default uses `errors.Is`. Override globally for custom error types like `serror.Trait`:

```go
testingx.SetDefaultErrMatcher(func(got error, want any) bool {
    return serror.OfTrait(got, want.(serror.Trait))
})
```

---

### CheckPanic — panic assertion

Runs `fn`, catches the expected panic, re-panics on anything unexpected.

```go
if tx.CheckPanic(func() {
    result = DoSomething(arg)
}, ErrExpectedPanic) {
    return // or continue in table test
}
```

Returns `true` if the expected panic occurred, `false` if fn ran cleanly.

---

### Low-level utils

For cases where the higher-level helpers create friction, use these directly:

```go
tx.TypeMismatch(got, (*MyStruct)(nil))   // got type *Foo, want *MyStruct
tx.ErrUnexpected(gotErr)                 // unexpected error: ...
tx.ErrExpected(wantErr)                  // expected error ..., got nil
tx.ErrMismatch(gotErr, wantErr)          // got error ..., want error ...
tx.Mismatch(diff)                        // mismatch (-want +got): ...
```

Useful when asserting concrete types before further checks:

```go
got, ok := gotE.(*ErrorRecord)
if !ok {
    tx.TypeMismatch(gotE, (*ErrorRecord)(nil))
    return
}
```

---

## Full example — table test

```go
func TestNew(t *testing.T) {
    tests := []struct {
        name       string
        input      string
        want       *Result
        wantErr    error
        wantPanic  any
    }{
        // ...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tx := testingx.New(t)

            var got *Result
            if tx.CheckPanic(func() {
                got, err = New(tt.input)
            }, tt.wantPanic) {
                return
            }

            tx.CheckErr(got, tt.want, err, tt.wantErr)
        })
    }
}
```
