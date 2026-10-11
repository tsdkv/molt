package molt_test

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/tsdkv/molt"
	"github.com/tsdkv/molt/internal/testutil"
	"github.com/tsdkv/molt/xf"
	"pgregory.net/rapid"
)

func sortInts(in []int) []int {
	out := slices.Clone(in)
	slices.Sort(out)
	return out
}

var sortMRs = []molt.MR[[]int, []int]{
	molt.Rel("reverse-invariant", xf.Reverse[int](), molt.EqualFunc[[]int](slices.Equal)),
	molt.RelCase("append-max-extends-output",
		molt.Pure(func(in []int) []int { return append(slices.Clone(in), slices.Max(in)+1) }),
		func(c molt.Case[[]int, []int]) error {
			if len(c.FollowUpOut) != len(c.SourceOut)+1 {
				return fmt.Errorf("follow-up output has %d elements, want %d",
					len(c.FollowUpOut), len(c.SourceOut)+1)
			}
			if !slices.Equal(c.SourceOut, c.FollowUpOut[:len(c.SourceOut)]) {
				return fmt.Errorf("follow-up output is not the source output extended")
			}
			return nil
		}).When(func(in []int) bool { return len(in) > 0 && slices.Max(in) < math.MaxInt }),
	molt.RelCase("output-is-permutation-of-input", molt.Identity[[]int](),
		func(c molt.Case[[]int, []int]) error {
			if !slices.Equal(sortInts(c.Source), sortInts(c.SourceOut)) {
				return fmt.Errorf("output is not a permutation of input")
			}
			return nil
		}),
}

func TestSortRelations(t *testing.T) {
	molt.Test(t, rapid.SliceOf(rapid.Int()), sortInts, sortMRs...)
}

func TestSortRelationsInsideRapid(t *testing.T) {
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
	tb := &testutil.TB{}
	rapid.Check(tb, func(t *rapid.T) {
		molt.Run(t, rapid.SliceOf(rapid.Int()), buggy, sortMRs...)
	})
	if !tb.Failed() {
		t.Fatal("buggy sort was not caught")
	}
	if !strings.Contains(tb.Output(), "metamorphic relation") {
		t.Fatalf("failure message does not mention the relation:\n%s", tb.Output())
	}
}

func TestRunExamples(t *testing.T) {
	examples := [][]int{{}, {1}, {math.MaxInt, 0}, {3, 1, 2}}
	t.Run("correct", func(t *testing.T) {
		molt.RunExamples(t, examples, sortInts, sortMRs...)
	})
	t.Run("mutation", func(t *testing.T) {
		buggy := func(in []int) []int {
			out := sortInts(in)
			if len(out) > 0 {
				out = append(out, out[0])
			}
			return out
		}
		tb := &testutil.TB{}
		molt.RunExamples(tb, examples, buggy, sortMRs...)
		if !tb.Failed() {
			t.Fatal("buggy sort was not caught on examples")
		}
	})
}

// Drawing transforms cannot run on fixed examples; the runner must say so
// instead of panicking on a nil *rapid.T.
func TestRunExamplesRejectsDrawing(t *testing.T) {
	tb := &testutil.TB{}
	molt.RunExamples(tb, [][]int{{1, 2}}, sortInts,
		molt.Rel("appends", xf.AppendRandom(rapid.Int()), molt.EqualFunc[[]int](slices.Equal)))
	if !tb.Failed() {
		t.Fatal("drawing transform was not rejected")
	}
	if !strings.Contains(tb.Output(), "cannot run on fixed examples") {
		t.Fatalf("unexpected message:\n%s", tb.Output())
	}
}

// Multi-input relations are ordinary relations over slices of inputs drawn
// with rapid.SliceOfN: here, commutativity of addition.
func TestCommutativity(t *testing.T) {
	sum := func(xs []int) int { return xs[0] + xs[1] }
	swap := molt.Pure(func(xs []int) []int { return []int{xs[1], xs[0]} })
	mrs := []molt.MR[[]int, int]{
		molt.Rel("commutative", swap, molt.Equal[int]()),
	}

	t.Run("correct", func(t *testing.T) {
		molt.Test(t, rapid.SliceOfN(rapid.Int(), 2, 2), sum, mrs...)
	})
	t.Run("mutation", func(t *testing.T) {
		tb := &testutil.TB{}
		rapid.Check(tb, func(t *rapid.T) {
			molt.Run(t, rapid.SliceOfN(rapid.Int(), 2, 2),
				func(xs []int) int { return xs[0] - xs[1] }, mrs...)
		})
		if !tb.Failed() {
			t.Fatal("non-commutative function was not caught")
		}
	})
}
