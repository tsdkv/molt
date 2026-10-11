package laws

import (
	"cmp"
	"fmt"
	"slices"
	"testing"

	"github.com/tsdkv/molt"
	"github.com/tsdkv/molt/xf"
	"pgregory.net/rapid"
)

// Set checks the laws of a set implementation. newSet builds a set from a
// list of elements; elements must return the set's contents in any order.
// union and inter may be nil, in which case the algebra laws are skipped
// and only the construction laws are checked.
//
// Construction laws: building a set is order- and duplication-invariant,
// and the result contains exactly the input elements. Algebra laws: union
// and intersection are commutative and idempotent, intersection is a subset
// of both operands, union a superset.
func Set[S any, E cmp.Ordered](
	t *testing.T,
	elemGen *rapid.Generator[E],
	newSet func([]E) S,
	elements func(S) []E,
	union, inter func(a, b S) S,
) {
	t.Helper()
	canonSet := func(s S) []E {
		out := slices.Clone(elements(s))
		slices.Sort(out)
		return slices.Compact(out)
	}
	canonList := func(elems []E) []E {
		return canonSet(newSet(elems))
	}

	listGen := rapid.SliceOf(elemGen)
	molt.Test(t, listGen, canonList, constructionMRs[E]()...)

	if union == nil || inter == nil {
		return
	}

	molt.Test(t, rapid.SliceOfN(listGen, 2, 2),
		func(srcs [][]E) setPair[E] {
			a, b := newSet(srcs[0]), newSet(srcs[1])
			return setPair[E]{canonSet(union(a, b)), canonSet(inter(a, b))}
		},
		algebraMRs(union, inter, canonList)...)
}

// setPair is the observable output of the set-algebra SUT: the canonical
// element lists of union(a, b) and intersection(a, b).
type setPair[E cmp.Ordered] struct {
	union []E
	inter []E
}

func constructionMRs[E cmp.Ordered]() []molt.MR[[]E, []E] {
	return []molt.MR[[]E, []E]{
		molt.Rel("construction-order-invariant", xf.Reverse[E](), molt.EqualFunc[[]E](slices.Equal)),
		molt.Rel("construction-dedup-invariant", xf.Duplicate[E](), molt.EqualFunc[[]E](slices.Equal)),
		molt.RelCase("construction-contains-all-inputs", molt.Identity[[]E](),
			func(c molt.Case[[]E, []E]) error {
				for _, e := range c.Source {
					if !slices.Contains(c.SourceOut, e) {
						return fmt.Errorf("set built from input does not contain %v", e)
					}
				}
				return nil
			}),
		molt.RelCase("construction-contains-nothing-else", molt.Identity[[]E](),
			func(c molt.Case[[]E, []E]) error {
				for _, e := range c.SourceOut {
					if !slices.Contains(c.Source, e) {
						return fmt.Errorf("set contains %v, which was never added", e)
					}
				}
				return nil
			}),
	}
}

func algebraMRs[S any, E cmp.Ordered](
	union, inter func(a, b S) S,
	canonList func([]E) []E,
) []molt.MR[[][]E, setPair[E]] {
	swap := molt.Pure(func(srcs [][]E) [][]E { return [][]E{srcs[1], srcs[0]} })
	keep := molt.Identity[[][]E]()
	return []molt.MR[[][]E, setPair[E]]{
		molt.Rel("union-commutative", swap,
			molt.Contra(func(p setPair[E]) []E { return p.union }, molt.EqualFunc[[]E](slices.Equal))),
		molt.Rel("intersection-commutative", swap,
			molt.Contra(func(p setPair[E]) []E { return p.inter }, molt.EqualFunc[[]E](slices.Equal))),
		molt.RelCase("union-idempotent",
			molt.Pure(func(srcs [][]E) [][]E { return [][]E{srcs[0], srcs[0]} }),
			func(c molt.Case[[][]E, setPair[E]]) error {
				if want := canonList(c.Source[0]); !slices.Equal(c.FollowUpOut.union, want) {
					return fmt.Errorf("union(a, a) = %v, want %v", c.FollowUpOut.union, want)
				}
				return nil
			}),
		molt.RelCase("intersection-subset", keep,
			func(c molt.Case[[][]E, setPair[E]]) error {
				for _, e := range c.SourceOut.inter {
					if !slices.Contains(c.Source[0], e) || !slices.Contains(c.Source[1], e) {
						return fmt.Errorf("intersection contains %v, absent from an operand", e)
					}
				}
				return nil
			}),
		molt.RelCase("union-superset", keep,
			func(c molt.Case[[][]E, setPair[E]]) error {
				for _, src := range c.Source {
					for _, e := range src {
						if !slices.Contains(c.SourceOut.union, e) {
							return fmt.Errorf("union misses %v from an operand", e)
						}
					}
				}
				return nil
			}),
	}
}
