package laws_test

import (
	"fmt"
	"strings"
)

// fakeTB is a minimal testing.TB that records failures, used to assert that
// law sets detect broken implementations.
type fakeTB struct {
	failed bool
	log    strings.Builder
}

func (f *fakeTB) Helper()                         {}
func (f *fakeTB) Name() string                    { return "fakeTB" }
func (f *fakeTB) Logf(format string, args ...any) { fmt.Fprintf(&f.log, format+"\n", args...) }
func (f *fakeTB) Log(args ...any)                 { fmt.Fprintln(&f.log, args...) }
func (f *fakeTB) Skipf(string, ...any)            {}
func (f *fakeTB) Skip(...any)                     {}
func (f *fakeTB) SkipNow()                        {}
func (f *fakeTB) Fail()                           { f.failed = true }
func (f *fakeTB) Failed() bool                    { return f.failed }
func (f *fakeTB) FailNow()                        { f.failed = true }
func (f *fakeTB) Error(args ...any)               { f.failed = true; f.Log(args...) }
func (f *fakeTB) Errorf(format string, args ...any) {
	f.failed = true
	f.Logf(format, args...)
}
func (f *fakeTB) Fatal(args ...any) { f.failed = true; f.Log(args...) }
func (f *fakeTB) Fatalf(format string, args ...any) {
	f.failed = true
	f.Logf(format, args...)
}
