package laws

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/tsdkv/molt"
	"pgregory.net/rapid"
)

// Sortable checks the standard contract for a sort function: the output is
// sorted, is a permutation of the input, does not depend on input order, and
// sorting it again changes nothing.
func Sortable[E cmp.Ordered](t *rapid.T, gen *rapid.Generator[[]E], sort func([]E) []E) {
	molt.Run(t, gen, sort,
		molt.MR[[]E, []E]{
			Name:      "output-sorted",
			Transform: molt.Identity[[]E],
			Check: func(c molt.Case[[]E, []E]) error {
				if !slices.IsSorted(c.SourceOut) {
					return fmt.Errorf("output is not sorted")
				}
				return nil
			},
		},
		molt.MR[[]E, []E]{
			Name:      "output-is-permutation-of-input",
			Transform: molt.Identity[[]E],
			Check: func(c molt.Case[[]E, []E]) error {
				counts := make(map[E]int, len(c.Source))
				for _, e := range c.Source {
					counts[e]++
				}
				for _, e := range c.SourceOut {
					counts[e]--
				}
				for _, n := range counts {
					if n != 0 {
						return fmt.Errorf("output is not a permutation of input")
					}
				}
				return nil
			},
		},
		molt.MR[[]E, []E]{
			Name: "reverse-invariant",
			Transform: func(_ *rapid.T, in []E) []E {
				out := slices.Clone(in)
				slices.Reverse(out)
				return out
			},
			Check: molt.EqualFunc[[]E, []E](slices.Equal),
		},
		molt.MR[[]E, []E]{
			Name:      "idempotent",
			Transform: molt.Identity[[]E],
			Check: func(c molt.Case[[]E, []E]) error {
				if !slices.Equal(c.SourceOut, sort(c.SourceOut)) {
					return fmt.Errorf("sorting the output changed it")
				}
				return nil
			},
		},
	)
}
