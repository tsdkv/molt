package molt_test

import (
	"math"
	"slices"
	"testing"

	"github.com/tsdkv/molt"
	"github.com/tsdkv/molt/internal/testutil"
	"pgregory.net/rapid"
)

var sortExamples = [][]int{
	{},
	{1},
	{2, 1},
	{5, 5, 5, 5},
	{math.MaxInt, math.MinInt, 0},
}

func TestRunExamples(t *testing.T) {
	// plain *testing.T — no rapid involved
	molt.RunExamples(t, sortExamples, sortInts, sortMRs...)
}

func TestRunExamplesMutationCaught(t *testing.T) {
	buggy := func(in []int) []int {
		out := sortInts(in)
		if len(out) > 0 {
			out = append(out, out[0])
		}
		return out
	}
	tb := &testutil.TB{}
	molt.RunExamples(tb, [][]int{{3, 1, 2}}, buggy, sortMRs...)
	if !tb.Failed() {
		t.Fatal("buggy sort was not caught over examples")
	}
}

func TestRunExamplesN(t *testing.T) {
	max := func(xs []int) int { return slices.Max(xs) }
	commutative := molt.MRN[int, int]{
		Name: "max-is-commutative",
		Transform: func(_ *rapid.T, srcs []int) []int {
			return []int{srcs[1], srcs[0]}
		},
		Check: func(c molt.CaseN[int, int]) error {
			return molt.Equal[int, int]()(
				molt.Case[int, int]{SourceOut: c.SourceOut, FollowUpOut: c.FollowUpOut})
		},
	}
	molt.RunExamplesN(t, [][]int{{1, 2}, {-5, 0}, {math.MaxInt, math.MinInt}}, max, commutative)
}
