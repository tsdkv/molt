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

// Sortable checks the standard contract for a sort function: the output is
// sorted, is a permutation of the input, does not depend on input order,
// and sorting it again changes nothing.
func Sortable[E cmp.Ordered](t *testing.T, gen *rapid.Generator[[]E], sort func([]E) []E) {
	t.Helper()
	molt.Test(t, gen, sort, sortableMRs(sort)...)
}

func sortableMRs[E cmp.Ordered](sort func([]E) []E) []molt.MR[[]E, []E] {
	return []molt.MR[[]E, []E]{
		molt.Rel("output-sorted", molt.Identity[[]E](),
			molt.Contra(func(out []E) bool { return slices.IsSorted(out) }, molt.Equal[bool]())),
		molt.RelCase("output-is-permutation-of-input", molt.Identity[[]E](),
			func(c molt.Case[[]E, []E]) error {
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
			}),
		molt.Rel("reverse-invariant", xf.Reverse[E](), molt.EqualFunc[[]E](slices.Equal)),
		molt.RelCase("idempotent", molt.Identity[[]E](),
			func(c molt.Case[[]E, []E]) error {
				if !slices.Equal(c.SourceOut, sort(c.SourceOut)) {
					return fmt.Errorf("sorting the output changed it")
				}
				return nil
			}),
	}
}
