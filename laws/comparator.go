package laws

import (
	"fmt"
	"testing"

	"github.com/tsdkv/molt"
	"pgregory.net/rapid"
)

// Comparator checks the standard contract for a three-way comparison
// function (negative / zero / positive): reflexive, antisymmetric,
// consistent with eq, and transitive.
//
// Multi-input relations (swap two inputs for antisymmetry, three inputs for
// transitivity) are ordinary relations over slices of inputs, drawn with
// rapid.SliceOfN.
func Comparator[T comparable](t *testing.T, gen *rapid.Generator[T], cmp func(a, b T) int, eq func(a, b T) bool) {
	t.Helper()
	molt.Test(t, gen, func(x T) int { return cmp(x, x) }, reflexiveMR[T]())
	molt.Test(t, rapid.SliceOfN(gen, 2, 2), func(xs []T) int { return cmp(xs[0], xs[1]) }, pairMRs(eq)...)
	molt.Test(t, rapid.SliceOfN(gen, 3, 3),
		func(xs []T) [3]int { return [3]int{cmp(xs[0], xs[1]), cmp(xs[1], xs[2]), cmp(xs[0], xs[2])} },
		transitiveMR[T]())
}

func reflexiveMR[T comparable]() molt.MR[T, int] {
	return molt.Rel("reflexive", molt.Identity[T](), func(srcOut, _ int) error {
		if srcOut != 0 {
			return fmt.Errorf("cmp(x, x) = %d, want 0", srcOut)
		}
		return nil
	})
}

func pairMRs[T comparable](eq func(a, b T) bool) []molt.MR[[]T, int] {
	swap := molt.Pure(func(xs []T) []T { return []T{xs[1], xs[0]} })
	return []molt.MR[[]T, int]{
		molt.Rel("antisymmetric", swap, func(srcOut, fupOut int) error {
			if sign(srcOut) != -sign(fupOut) {
				return fmt.Errorf("sign(cmp(a, b)) = %d, sign(cmp(b, a)) = %d", sign(srcOut), sign(fupOut))
			}
			return nil
		}),
		molt.RelCase("consistent-with-equal", molt.Identity[[]T](), func(c molt.Case[[]T, int]) error {
			if (c.SourceOut == 0) != eq(c.Source[0], c.Source[1]) {
				return fmt.Errorf("cmp(a, b) == 0 disagrees with eq(a, b)")
			}
			return nil
		}),
	}
}

func transitiveMR[T comparable]() molt.MR[[]T, [3]int] {
	return molt.Rel("transitive", molt.Identity[[]T](), func(srcOut, _ [3]int) error {
		ab, bc, ac := srcOut[0], srcOut[1], srcOut[2]
		if ab <= 0 && bc <= 0 && ac > 0 {
			return fmt.Errorf("cmp(a,b)=%d, cmp(b,c)=%d but cmp(a,c)=%d", ab, bc, ac)
		}
		return nil
	})
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
