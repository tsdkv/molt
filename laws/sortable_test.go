package laws_test

import (
	"slices"
	"testing"

	"github.com/tsdkv/molt/internal/testutil"
	"github.com/tsdkv/molt/laws"
	"pgregory.net/rapid"
)

func sortInts(in []int) []int {
	out := slices.Clone(in)
	slices.Sort(out)
	return out
}

func TestSortable(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		laws.Sortable(t, rapid.SliceOf(rapid.Int()), sortInts)
	})
}

// A sort that drops the minimum element must be caught by the
// output-is-permutation-of-input law.
func TestSortableMutationCaught(t *testing.T) {
	buggy := func(in []int) []int {
		out := sortInts(in)
		if len(out) > 0 {
			out = out[1:]
		}
		return out
	}
	tb := &testutil.TB{}
	rapid.Check(tb, func(t *rapid.T) {
		laws.Sortable(t, rapid.SliceOf(rapid.Int()), buggy)
	})
	if !tb.Failed() {
		t.Fatal("element-dropping sort was not caught")
	}
}
