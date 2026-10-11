package molt_test

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/tsdkv/molt"
	"pgregory.net/rapid"
)

func sortInts(in []int) []int {
	out := slices.Clone(in)
	slices.Sort(out)
	return out
}

var sortMRs = []molt.MR[[]int, []int]{
	{
		Name: "reverse-invariant",
		Transform: func(_ *rapid.T, in []int) []int {
			out := slices.Clone(in)
			slices.Reverse(out)
			return out
		},
		Check: molt.EqualFunc[[]int, []int](slices.Equal),
	},
	{
		Name: "append-max-extends-output",
		When: func(in []int) bool { return len(in) > 0 && slices.Max(in) < math.MaxInt },
		Transform: func(_ *rapid.T, in []int) []int {
			return append(slices.Clone(in), slices.Max(in)+1)
		},
		Check: func(c molt.Case[[]int, []int]) error {
			if len(c.FollowUpOut) != len(c.SourceOut)+1 {
				return fmt.Errorf("follow-up output has %d elements, want %d",
					len(c.FollowUpOut), len(c.SourceOut)+1)
			}
			if !slices.Equal(c.SourceOut, c.FollowUpOut[:len(c.SourceOut)]) {
				return fmt.Errorf("follow-up output is not the source output extended")
			}
			return nil
		},
	},
	{
		Name:      "output-is-permutation-of-input",
		Transform: func(_ *rapid.T, in []int) []int { return slices.Clone(in) },
		Check: func(c molt.Case[[]int, []int]) error {
			want := sortInts(c.Source)
			got := sortInts(c.SourceOut)
			if !slices.Equal(want, got) {
				return fmt.Errorf("output is not a permutation of input")
			}
			return nil
		},
	},
}

func TestSortRelations(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		molt.Run(t, rapid.SliceOf(rapid.Int()), sortInts, sortMRs...)
	})
}

// A broken sort that duplicates the minimum element satisfies
// reverse-invariance but violates the other relations: the suite must catch
// what any single relation cannot.
func TestSortMutationCaught(t *testing.T) {
	buggy := func(in []int) []int {
		out := sortInts(in)
		if len(out) > 0 {
			out = append(out, out[0])
		}
		return out
	}
	tb := &fakeTB{}
	rapid.Check(tb, func(t *rapid.T) {
		molt.Run(t, rapid.SliceOf(rapid.Int()), buggy, sortMRs...)
	})
	if !tb.Failed() {
		t.Fatal("buggy sort was not caught")
	}
	if !strings.Contains(tb.log.String(), "metamorphic relation") {
		t.Fatalf("failure message does not mention the relation:\n%s", tb.log.String())
	}
}
