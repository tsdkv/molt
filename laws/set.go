package laws

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/tsdkv/molt"
	"pgregory.net/rapid"
)

// Set checks the laws of a set implementation. newSet builds a set from a
// list of elements; elements must return the set's contents in any order.
// union and inter may be nil, in which case the algebra laws are skipped and
// only the construction laws are checked.
//
// Construction laws: building a set is order- and duplication-invariant, and
// the result contains exactly the input elements. Algebra laws: union and
// intersection are commutative and idempotent, intersection is a subset of
// both operands, union a superset.
func Set[S any, E cmp.Ordered](
	t *rapid.T,
	elemGen *rapid.Generator[E],
	newSet func([]E) S,
	elements func(S) []E,
	union, inter func(a, b S) S,
) {
	canonSet := func(s S) []E {
		out := slices.Clone(elements(s))
		slices.Sort(out)
		return slices.Compact(out)
	}
	canonList := func(elems []E) []E {
		return canonSet(newSet(elems))
	}

	listGen := rapid.SliceOf(elemGen)
	molt.Run(t, listGen, canonList,
		molt.MR[[]E, []E]{
			Name: "construction-order-invariant",
			Transform: func(_ *rapid.T, in []E) []E {
				out := slices.Clone(in)
				slices.Reverse(out)
				return out
			},
			Check: molt.EqualFunc[[]E, []E](slices.Equal),
		},
		molt.MR[[]E, []E]{
			Name: "construction-dedup-invariant",
			Transform: func(_ *rapid.T, in []E) []E {
				return append(slices.Clone(in), in...)
			},
			Check: molt.EqualFunc[[]E, []E](slices.Equal),
		},
		molt.MR[[]E, []E]{
			Name:      "construction-contains-all-inputs",
			Transform: molt.Identity[[]E],
			Check: func(c molt.Case[[]E, []E]) error {
				for _, e := range c.Source {
					if !slices.Contains(c.SourceOut, e) {
						return fmt.Errorf("set built from input does not contain %v", e)
					}
				}
				return nil
			},
		},
		molt.MR[[]E, []E]{
			Name:      "construction-contains-nothing-else",
			Transform: molt.Identity[[]E],
			Check: func(c molt.Case[[]E, []E]) error {
				for _, e := range c.SourceOut {
					if !slices.Contains(c.Source, e) {
						return fmt.Errorf("set contains %v, which was never added", e)
					}
				}
				return nil
			},
		},
	)

	if union == nil || inter == nil {
		return
	}

	type pair struct {
		union []E
		inter []E
	}
	f2 := func(srcs [][]E) pair {
		a, b := newSet(srcs[0]), newSet(srcs[1])
		return pair{canonSet(union(a, b)), canonSet(inter(a, b))}
	}
	keep := func(_ *rapid.T, srcs [][]E) [][]E { return srcs }
	swap := func(_ *rapid.T, srcs [][]E) [][]E { return [][]E{srcs[1], srcs[0]} }

	molt.RunN(t, listGen, 2, f2,
		molt.MRN[[]E, pair]{
			Name:      "union-commutative",
			Transform: swap,
			Check: func(c molt.CaseN[[]E, pair]) error {
				if !slices.Equal(c.SourceOut.union, c.FollowUpOut.union) {
					return fmt.Errorf("union(a, b) != union(b, a)")
				}
				return nil
			},
		},
		molt.MRN[[]E, pair]{
			Name:      "intersection-commutative",
			Transform: swap,
			Check: func(c molt.CaseN[[]E, pair]) error {
				if !slices.Equal(c.SourceOut.inter, c.FollowUpOut.inter) {
					return fmt.Errorf("inter(a, b) != inter(b, a)")
				}
				return nil
			},
		},
		molt.MRN[[]E, pair]{
			Name: "union-idempotent",
			Transform: func(_ *rapid.T, srcs [][]E) [][]E {
				return [][]E{srcs[0], srcs[0]}
			},
			Check: func(c molt.CaseN[[]E, pair]) error {
				if want := canonList(c.Sources[0]); !slices.Equal(c.FollowUpOut.union, want) {
					return fmt.Errorf("union(a, a) = %v, want %v", c.FollowUpOut.union, want)
				}
				return nil
			},
		},
		molt.MRN[[]E, pair]{
			Name:      "intersection-subset",
			Transform: keep,
			Check: func(c molt.CaseN[[]E, pair]) error {
				for _, e := range c.SourceOut.inter {
					if !slices.Contains(c.Sources[0], e) || !slices.Contains(c.Sources[1], e) {
						return fmt.Errorf("intersection contains %v, absent from an operand", e)
					}
				}
				return nil
			},
		},
		molt.MRN[[]E, pair]{
			Name:      "union-superset",
			Transform: keep,
			Check: func(c molt.CaseN[[]E, pair]) error {
				for _, src := range c.Sources {
					for _, e := range src {
						if !slices.Contains(c.SourceOut.union, e) {
							return fmt.Errorf("union misses %v from an operand", e)
						}
					}
				}
				return nil
			},
		},
	)
}
