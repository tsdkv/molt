package molt_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/tsdkv/molt"
	"github.com/tsdkv/molt/internal/testutil"
	"pgregory.net/rapid"
)

// Commutativity: two independent sources, the transform swaps them.
func TestRunNCommutativity(t *testing.T) {
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
	rapid.Check(t, func(t *rapid.T) {
		molt.RunN(t, rapid.Int(), 2, max, commutative)
	})
}

// A "max" that actually returns the first argument is not commutative and
// must be caught.
func TestRunNCommutativityMutationCaught(t *testing.T) {
	first := func(xs []int) int { return xs[0] }
	commutative := molt.MRN[int, int]{
		Name: "is-commutative",
		Transform: func(_ *rapid.T, srcs []int) []int {
			return []int{srcs[1], srcs[0]}
		},
		Check: func(c molt.CaseN[int, int]) error {
			return molt.Equal[int, int]()(
				molt.Case[int, int]{SourceOut: c.SourceOut, FollowUpOut: c.FollowUpOut})
		},
	}
	tb := &testutil.TB{}
	rapid.Check(tb, func(t *rapid.T) {
		molt.RunN(t, rapid.Int(), 2, first, commutative)
	})
	if !tb.Failed() {
		t.Fatal("non-commutative function was not caught")
	}
}

// Conditional relations over three sources: transitivity of <=, expressed
// with an identity transform and a check that only bites when the premise
// holds.
func TestRunNTransitivity(t *testing.T) {
	leq := func(xs []int) bool { return xs[0] <= xs[1] }
	transitive := molt.MRN[int, bool]{
		Name:      "leq-is-transitive",
		Transform: molt.Identity[[]int],
		Check: func(c molt.CaseN[int, bool]) error {
			a, b, d := c.Sources[0], c.Sources[1], c.Sources[2]
			if a <= b && b <= d && !(a <= d) {
				return errors.New("transitivity violated")
			}
			return nil
		},
	}
	rapid.Check(t, func(t *rapid.T) {
		molt.RunN(t, rapid.Int(), 3, leq, transitive)
	})
}
