// Package testutil provides helpers for molt's own tests.
package testutil

import (
	"fmt"
	"strings"
)

// TB is a minimal implementation of rapid's testing interface that records
// failures instead of aborting the test, letting tests assert that a
// relation or law set detects a deliberately broken implementation.
type TB struct {
	failed bool
	log    strings.Builder
}

// Output returns everything logged so far.
func (t *TB) Output() string { return t.log.String() }

func (t *TB) Helper()                         {}
func (t *TB) Name() string                    { return "testutil.TB" }
func (t *TB) Logf(format string, args ...any) { fmt.Fprintf(&t.log, format+"\n", args...) }
func (t *TB) Log(args ...any)                 { fmt.Fprintln(&t.log, args...) }
func (t *TB) Skipf(string, ...any)            {}
func (t *TB) Skip(...any)                     {}
func (t *TB) SkipNow()                        {}
func (t *TB) Fail()                           { t.failed = true }
func (t *TB) Failed() bool                    { return t.failed }
func (t *TB) FailNow()                        { t.failed = true }
func (t *TB) Error(args ...any)               { t.failed = true; t.Log(args...) }
func (t *TB) Errorf(format string, args ...any) {
	t.failed = true
	t.Logf(format, args...)
}
func (t *TB) Fatal(args ...any) { t.failed = true; t.Log(args...) }
func (t *TB) Fatalf(format string, args ...any) {
	t.failed = true
	t.Logf(format, args...)
}
