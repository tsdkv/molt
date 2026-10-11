// Package xf provides ready-made input transformations for molt — the
// transform half of a metamorphic relation. Apply them to a field of a
// larger input with molt.On, and compose them with molt.Compose and
// molt.PowerSet.
package xf

import (
	"math/rand/v2"
	"slices"

	"github.com/tsdkv/molt"
	"pgregory.net/rapid"
)

// Reverse returns a pure transform that reverses a slice.
func Reverse[E any]() molt.Tx[[]E] {
	return molt.Pure(func(in []E) []E {
		out := slices.Clone(in)
		slices.Reverse(out)
		return out
	})
}

// Shuffle returns a transform that randomly permutes a slice. The
// permutation is seeded by a rapid draw, so it is reproducible and
// shrinkable like any other input.
func Shuffle[E any]() molt.Tx[[]E] {
	return molt.Drawn(rapid.Int64(), func(in []E, seed int64) []E {
		out := slices.Clone(in)
		rand.New(rand.NewPCG(uint64(seed), 0)).Shuffle(len(out), func(i, j int) {
			out[i], out[j] = out[j], out[i]
		})
		return out
	})
}

// AppendRandom returns a transform that appends one element drawn from gen.
func AppendRandom[E any](gen *rapid.Generator[E]) molt.Tx[[]E] {
	return molt.Drawn(gen, func(in []E, e E) []E {
		return append(slices.Clone(in), e)
	})
}

// DropRandom returns a transform that removes one element at a drawn
// position. Draws cannot depend on the input, so the position is drawn as
// an arbitrary uint64 and reduced modulo the length.
func DropRandom[E any]() molt.Tx[[]E] {
	return molt.Drawn(rapid.Uint64(), func(in []E, i uint64) []E {
		if len(in) == 0 {
			return in
		}
		pos := int(i % uint64(len(in)))
		return slices.Delete(slices.Clone(in), pos, pos+1)
	})
}

// Duplicate returns a pure transform that concatenates a slice with itself
// (construction-dedup invariance, double-application, ...).
func Duplicate[E any]() molt.Tx[[]E] {
	return molt.Pure(func(in []E) []E {
		return append(slices.Clone(in), in...)
	})
}

// AddConst returns a pure transform shifting a float by a constant.
func AddConst(c float64) molt.Tx[float64] {
	return molt.Pure(func(x float64) float64 { return x + c })
}

// Negate returns a pure transform negating a float.
func Negate() molt.Tx[float64] {
	return molt.Pure(func(x float64) float64 { return -x })
}
