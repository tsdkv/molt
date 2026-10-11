package laws

import (
	"fmt"

	"github.com/tsdkv/molt"
	"pgregory.net/rapid"
)

// Comparator checks the standard contract for a three-way comparison
// function (negative / zero / positive): reflexive, antisymmetric,
// consistent with eq, and transitive.
func Comparator[T comparable](t *rapid.T, gen *rapid.Generator[T], cmp func(a, b T) int, eq func(a, b T) bool) {
	keep := func(_ *rapid.T, srcs []T) []T { return srcs }
	swap := func(_ *rapid.T, srcs []T) []T { return []T{srcs[1], srcs[0]} }

	molt.RunN(t, gen, 1,
		func(xs []T) int { return cmp(xs[0], xs[0]) },
		molt.MRN[T, int]{
			Name:      "reflexive",
			Transform: keep,
			Check: func(c molt.CaseN[T, int]) error {
				if c.SourceOut != 0 {
					return fmt.Errorf("cmp(x, x) = %d, want 0", c.SourceOut)
				}
				return nil
			},
		},
	)

	molt.RunN(t, gen, 2,
		func(xs []T) int { return cmp(xs[0], xs[1]) },
		molt.MRN[T, int]{
			Name:      "antisymmetric",
			Transform: swap,
			Check: func(c molt.CaseN[T, int]) error {
				if sign(c.SourceOut) != -sign(c.FollowUpOut) {
					return fmt.Errorf("sign(cmp(a, b)) = %d, sign(cmp(b, a)) = %d",
						sign(c.SourceOut), sign(c.FollowUpOut))
				}
				return nil
			},
		},
		molt.MRN[T, int]{
			Name:      "consistent-with-equal",
			Transform: keep,
			Check: func(c molt.CaseN[T, int]) error {
				if (c.SourceOut == 0) != eq(c.Sources[0], c.Sources[1]) {
					return fmt.Errorf("cmp(a, b) == 0 disagrees with eq(a, b)")
				}
				return nil
			},
		},
	)

	molt.RunN(t, gen, 3,
		func(xs []T) [3]int {
			return [3]int{cmp(xs[0], xs[1]), cmp(xs[1], xs[2]), cmp(xs[0], xs[2])}
		},
		molt.MRN[T, [3]int]{
			Name:      "transitive",
			Transform: keep,
			Check: func(c molt.CaseN[T, [3]int]) error {
				ab, bc, ac := c.SourceOut[0], c.SourceOut[1], c.SourceOut[2]
				if ab <= 0 && bc <= 0 && ac > 0 {
					return fmt.Errorf("cmp(a,b)=%d, cmp(b,c)=%d but cmp(a,c)=%d", ab, bc, ac)
				}
				return nil
			},
		},
	)
}

func sign(x int) int {
	switch {
	case x < 0:
		return -1
	case x > 0:
		return 1
	default:
		return 0
	}
}
