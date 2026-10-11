package molt

import (
	"strings"

	"pgregory.net/rapid"
)

// Identity is a transformation that returns its input unchanged. Useful when
// the interesting part of a relation lives entirely in the check (round-trip
// properties, invariants over a single run).
func Identity[I any](_ *rapid.T, in I) I { return in }

// Compose returns a transformation that applies first and then second.
// Any values the transformations draw from t are drawn in that order.
func Compose[I any](first, second func(*rapid.T, I) I) func(*rapid.T, I) I {
	return func(t *rapid.T, in I) I {
		return second(t, first(t, in))
	}
}

// NamedTransform pairs a name with a transformation, so that failures
// produced by PowerSet can say which combination of transformations broke
// the relation.
type NamedTransform[I any] struct {
	Name      string
	Transform func(t *rapid.T, in I) I
}

// PowerSet expands a single relation and a set of orthogonal transformations
// into one MR per non-empty combination of transformations: each one alone,
// each pair composed, and so on up to all of them composed. The resulting
// MRs are passed to Run like any others:
//
//	molt.Run(t, gen, f, molt.PowerSet("sin-laws", approx,
//		molt.NamedTransform[float64]{Name: "shift-2π", Transform: shift2Pi},
//		molt.NamedTransform[float64]{Name: "mirror-π", Transform: mirrorPi},
//	)...)
//
// Compositions are where independently-correct transformations interfere,
// and where a surprising share of bugs live.
func PowerSet[I, O any](name string, check Relation[I, O], trs ...NamedTransform[I]) []MR[I, O] {
	var mrs []MR[I, O]
	for mask := 1; mask < 1<<len(trs); mask++ {
		var names []string
		transform := Identity[I]
		for i, tr := range trs {
			if mask&(1<<i) == 0 {
				continue
			}
			names = append(names, tr.Name)
			transform = Compose(transform, tr.Transform)
		}
		mrs = append(mrs, MR[I, O]{
			Name:      name + " [" + strings.Join(names, " + ") + "]",
			Transform: transform,
			Check:     check,
		})
	}
	return mrs
}
