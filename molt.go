// Package molt provides metamorphic testing for Go on top of rapid.
//
// Metamorphic testing checks how the outputs of a program relate across
// runs with different but related inputs, instead of checking outputs
// against known-correct answers. It is useful when writing down the expected
// output is hard or impossible: complex data structures, encoders, queries,
// numerical code.
//
// A metamorphic relation pairs an input transformation ("reverse the list")
// with a relation that must hold between the two resulting outputs ("the
// sorted results are identical"). State relations as data with Rel and
// RelCase, then run them with Test:
//
//	func TestSort(t *testing.T) {
//		molt.Test(t, rapid.SliceOf(rapid.Int()), mysort,
//			molt.Rel("reverse-invariant", xf.Reverse[int](), molt.EqualFunc[[]int](slices.Equal)),
//		)
//	}
//
// Each relation runs as its own subtest with its own rapid session. Because
// all inputs are drawn through rapid, failing cases are reproducible from a
// seed and are automatically minimized: only the source input is drawn, and
// the follow-up input is re-derived on every iteration, so shrinking always
// keeps the two consistent.
package molt

import (
	"strings"
)

// Case is one full execution of a metamorphic relation: the source input,
// the follow-up input derived from it, and the outputs observed for both.
type Case[I, O any] struct {
	Source      I
	FollowUp    I
	SourceOut   O
	FollowUpOut O
}

// MR is a metamorphic relation stated as data. Build it with Rel (the check
// sees only the two outputs) or RelCase (the check sees the full Case), and
// optionally restrict it with When.
type MR[I, O any] struct {
	name  string
	when  func(I) bool
	tx    Tx[I]
	check func(Case[I, O]) error
}

// Rel builds a metamorphic relation from a transform and an output-only
// relation — the common case:
//
//	molt.Rel("reverse-invariant", xf.Reverse[int](), molt.EqualFunc[[]int](slices.Equal))
func Rel[I, O any](name string, tx Tx[I], rel Relation[O]) MR[I, O] {
	return MR[I, O]{
		name: name,
		tx:   tx,
		check: func(c Case[I, O]) error {
			return rel(c.SourceOut, c.FollowUpOut)
		},
	}
}

// RelCase is Rel with a check that receives the full Case — source input,
// follow-up input, and both outputs — for relations that need the inputs
// (round-trips, "output is a permutation of input", ...).
func RelCase[I, O any](name string, tx Tx[I], check func(Case[I, O]) error) MR[I, O] {
	return MR[I, O]{name: name, tx: tx, check: check}
}

// When restricts the relation to source inputs satisfying pred. The
// precondition is re-evaluated for every (possibly shrunk) candidate, and
// Test warns when it filters out most inputs.
func (m MR[I, O]) When(pred func(I) bool) MR[I, O] {
	m.when = pred
	return m
}

// Name returns the relation's name, as passed to Rel or RelCase.
func (m MR[I, O]) Name() string { return m.name }

// Named pairs a name with a transform, so that failures produced by
// PowerSet can say which combination of transforms broke the relation.
type Named[I any] struct {
	Name string
	Tx   Tx[I]
}

// PowerSet expands a single relation and a set of orthogonal transforms
// into one MR per non-empty combination of transforms: each one alone, each
// pair composed, and so on up to all of them composed. The resulting MRs
// are passed to Test or Run like any others:
//
//	mrs := molt.PowerSet("sin", molt.ApproxEqual(1e-9),
//		molt.Named[float64]{Name: "shift-2pi", Tx: xf.AddConst(2 * math.Pi)},
//		molt.Named[float64]{Name: "mirror-pi", Tx: molt.Pure(func(x float64) float64 { return math.Pi - x })},
//	)
//
// Compositions are where independently-correct transforms interfere, and
// where a surprising share of bugs live.
func PowerSet[I, O any](name string, rel Relation[O], trs ...Named[I]) []MR[I, O] {
	var mrs []MR[I, O]
	for mask := 1; mask < 1<<len(trs); mask++ {
		var names []string
		tx := Identity[I]()
		for i, tr := range trs {
			if mask&(1<<i) == 0 {
				continue
			}
			names = append(names, tr.Name)
			tx = Compose(tx, tr.Tx)
		}
		mrs = append(mrs, Rel(name+" ["+strings.Join(names, " + ")+"]", tx, rel))
	}
	return mrs
}
