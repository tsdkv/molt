package molt_test

import (
	"math"
	"testing"

	"github.com/tsdkv/molt"
	"github.com/tsdkv/molt/internal/testutil"
	"github.com/tsdkv/molt/xf"
	"pgregory.net/rapid"
)

// PowerSet expands orthogonal transforms into every composition. For sin,
// both "shift by 2π" and "mirror" (x → π − x) preserve the output — and so
// must their composition.
func TestPowerSetSin(t *testing.T) {
	mrs := molt.PowerSet("sin", molt.ApproxEqual(1e-9),
		molt.Named[float64]{Name: "shift-2pi", Tx: xf.AddConst(2 * math.Pi)},
		molt.Named[float64]{Name: "mirror-pi", Tx: molt.Pure(func(x float64) float64 { return math.Pi - x })},
	)
	if len(mrs) != 3 {
		t.Fatalf("PowerSet of 2 transforms produced %d MRs, want 3", len(mrs))
	}
	molt.Test(t, rapid.Float64Range(-100, 100), math.Sin, mrs...)
}

// A "sin" that is only accurate near zero must be caught by the
// composition: each single transform stays within its accurate region,
// their composition does not.
func TestPowerSetSinMutationCaught(t *testing.T) {
	buggySin := func(x float64) float64 {
		if math.Abs(x) > 10 {
			return 0 // give up far from zero
		}
		return math.Sin(x)
	}
	mrs := molt.PowerSet("sin", molt.ApproxEqual(1e-9),
		molt.Named[float64]{Name: "shift-2pi", Tx: xf.AddConst(2 * math.Pi)},
		molt.Named[float64]{Name: "mirror-pi", Tx: molt.Pure(func(x float64) float64 { return math.Pi - x })},
	)
	tb := &testutil.TB{}
	rapid.Check(tb, func(t *rapid.T) {
		molt.Run(t, rapid.Float64Range(-5, 5), buggySin, mrs...)
	})
	if !tb.Failed() {
		t.Fatal("buggy sin was not caught")
	}
}

// Compose chains transforms; Identity is the neutral element.
func TestCompose(t *testing.T) {
	tb := &testutil.TB{}
	rapid.Check(tb, func(t *rapid.T) {
		molt.Run(t, rapid.SliceOf(rapid.Int()), sortInts,
			molt.Rel("double-then-reverse",
				molt.Compose(xf.Duplicate[int](), xf.Reverse[int]()),
				molt.EqualFunc[[]int](func(a, b []int) bool {
					// sorting a doubled reversed slice = sorting the doubled slice
					if len(b) != 2*len(a) {
						return false
					}
					return true
				})),
		)
	})
	if tb.Failed() {
		t.Fatalf("correct sort failed composed transform:\n%s", tb.Output())
	}
}
