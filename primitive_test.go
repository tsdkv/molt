package molt_test

import (
	"slices"
	"testing"

	"github.com/tsdkv/molt"
	"github.com/tsdkv/molt/internal/testutil"
	"pgregory.net/rapid"
)

// The runners are conveniences; the primitives compose with free-form rapid
// tests directly.
func TestPrimitivesFreeForm(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		src, fup := molt.Follow(t, rapid.SliceOf(rapid.Int()), func(_ *rapid.T, in []int) []int {
			out := slices.Clone(in)
			slices.Reverse(out)
			return out
		})
		c := molt.Execute(sortInts, src, fup)
		molt.Assert(t, "reverse-invariant", c, molt.EqualFunc[[]int, []int](slices.Equal))
	})
}

func TestPrimitivesMutationCaught(t *testing.T) {
	buggy := func(in []int) []int {
		if len(in) > 0 {
			in = in[1:] // drops the first element of the input
		}
		return sortInts(in)
	}
	tb := &testutil.TB{}
	rapid.Check(tb, func(t *rapid.T) {
		src, fup := molt.Follow(t, rapid.SliceOf(rapid.Int()), func(_ *rapid.T, in []int) []int {
			out := slices.Clone(in)
			slices.Reverse(out)
			return out
		})
		molt.Assert(t, "reverse-invariant", molt.Execute(buggy, src, fup),
			molt.EqualFunc[[]int, []int](slices.Equal))
	})
	if !tb.Failed() {
		t.Fatal("buggy sort was not caught through the primitives")
	}
}
