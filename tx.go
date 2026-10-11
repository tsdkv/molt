package molt

import "pgregory.net/rapid"

// Tx is an input transformation: it derives the follow-up input from the
// source input. Construct one with Pure (no randomness), Drawn (declares
// the extra randomness it needs as a generator), or Func (escape hatch:
// draws from the rapid state directly). Ready-made transforms live in the
// xf package.
//
// Because the randomness a transform needs is declared up front, transforms
// never see math/rand and never see a nil *rapid.T; runners that cannot
// draw (RunExamples) reject drawing transforms with a clear message.
type Tx[I any] struct {
	draws    bool
	followUp func(t *rapid.T, in I) I
}

// Pure returns a transform that maps the input without extra randomness.
func Pure[I any](apply func(in I) I) Tx[I] {
	return Tx[I]{followUp: func(_ *rapid.T, in I) I { return apply(in) }}
}

// Identity is a transform that returns its input unchanged. Useful when the
// interesting part of a relation lives entirely in the check (round-trip
// properties, invariants over a single run).
func Identity[I any]() Tx[I] {
	return Pure(func(in I) I { return in })
}

// Drawn returns a transform that first draws auxiliary randomness from gen
// — reproducible and shrinkable like any rapid draw — and then applies it
// to the input. The draw cannot depend on the input value; normalize inside
// apply (e.g. an index modulo the length). R may itself be a struct drawn
// with rapid.Custom.
func Drawn[I, R any](gen *rapid.Generator[R], apply func(in I, r R) I) Tx[I] {
	return Tx[I]{
		draws: true,
		followUp: func(t *rapid.T, in I) I {
			return apply(in, gen.Draw(t, "transform"))
		},
	}
}

// Func adapts a transform that draws from the rapid state directly. Use it
// when Drawn is not expressive enough: draws that depend on the input,
// several draws, draws of different types.
func Func[I any](f func(t *rapid.T, in I) I) Tx[I] {
	return Tx[I]{draws: true, followUp: f}
}

// On lifts a transform on a part (a field, an element) to a transform on
// the whole: get extracts the part, set puts the transformed part back.
// This is how slice and number transforms apply to complex inputs:
//
//	molt.On(xf.Shuffle[string](),
//		func(c config) []string { return c.Tags },
//		func(c config, tags []string) config { c.Tags = tags; return c })
func On[S, I any](tx Tx[I], get func(S) I, set func(S, I) S) Tx[S] {
	return Tx[S]{
		draws: tx.draws,
		followUp: func(t *rapid.T, s S) S {
			return set(s, tx.followUp(t, get(s)))
		},
	}
}

// Compose returns a transform that applies first and then second. Any
// values the transforms draw are drawn in that order.
func Compose[I any](first, second Tx[I]) Tx[I] {
	return Tx[I]{
		draws: first.draws || second.draws,
		followUp: func(t *rapid.T, in I) I {
			return second.followUp(t, first.followUp(t, in))
		},
	}
}
