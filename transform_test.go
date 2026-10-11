package molt_test

import (
	"math"
	"slices"
	"testing"

	"github.com/tsdkv/molt"
	"pgregory.net/rapid"
)

// PowerSet expands orthogonal transformations into every composition. For
// sin, both "shift by 2π" and "mirror" (x → π − x) preserve the output — and
// so must their composition.
func TestPowerSetSin(t *testing.T) {
	shift2Pi := molt.NamedTransform[float64]{
		Name:      "shift-2pi",
		Transform: func(_ *rapid.T, x float64) float64 { return x + 2*math.Pi },
	}
	mirrorPi := molt.NamedTransform[float64]{
		Name:      "mirror-pi",
		Transform: func(_ *rapid.T, x float64) float64 { return math.Pi - x },
	}

	mrs := molt.PowerSet("sin", molt.ApproxEqual[float64](1e-9), shift2Pi, mirrorPi)
	if len(mrs) != 3 {
		t.Fatalf("PowerSet of 2 transforms produced %d MRs, want 3", len(mrs))
	}

	rapid.Check(t, func(t *rapid.T) {
		molt.Run(t, rapid.Float64Range(-100, 100), math.Sin, mrs...)
	})
}

// A "sin" that is only accurate near zero must be caught by the composition:
// each single transform stays within its accurate region, their composition
// does not.
func TestPowerSetSinMutationCaught(t *testing.T) {
	buggySin := func(x float64) float64 {
		if math.Abs(x) > 10 {
			return 0 // give up far from zero
		}
		return math.Sin(x)
	}
	shift2Pi := molt.NamedTransform[float64]{
		Name:      "shift-2pi",
		Transform: func(_ *rapid.T, x float64) float64 { return x + 2*math.Pi },
	}
	mirrorPi := molt.NamedTransform[float64]{
		Name:      "mirror-pi",
		Transform: func(_ *rapid.T, x float64) float64 { return math.Pi - x },
	}
	mrs := molt.PowerSet("sin", molt.ApproxEqual[float64](1e-9), shift2Pi, mirrorPi)

	tb := &fakeTB{}
	rapid.Check(tb, func(t *rapid.T) {
		molt.Run(t, rapid.Float64Range(-5, 5), buggySin, mrs...)
	})
	if !tb.Failed() {
		t.Fatal("buggy sin was not caught")
	}
}

// Transform combinators are ordinary functions and compose with Run.
func TestComposeWithRun(t *testing.T) {
	double := molt.Compose(
		molt.Identity[[]int],
		func(_ *rapid.T, in []int) []int { return append(slices.Clone(in), in...) },
	)
	mr := molt.MR[[]int, []int]{
		Name:      "doubling-preserves-max",
		Transform: double,
		Check: molt.Contra(
			func(out []int) int {
				if len(out) == 0 {
					return 0
				}
				return out[len(out)-1]
			},
			molt.Equal[[]int, int](),
		),
	}
	rapid.Check(t, func(t *rapid.T) {
		molt.Run(t, rapid.SliceOf(rapid.Int()), sortInts, mr)
	})
}
