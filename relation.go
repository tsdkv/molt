package molt

import (
	"cmp"
	"fmt"
	"maps"
	"math"
)

// Equal returns a Relation requiring both outputs to be ==.
func Equal[I any, O comparable]() Relation[I, O] {
	return func(c Case[I, O]) error {
		if c.SourceOut != c.FollowUpOut {
			return fmt.Errorf("outputs are not equal")
		}
		return nil
	}
}

// EqualFunc returns a Relation requiring both outputs to be equal according
// to eq (for example slices.Equal).
func EqualFunc[I, O any](eq func(a, b O) bool) Relation[I, O] {
	return func(c Case[I, O]) error {
		if !eq(c.SourceOut, c.FollowUpOut) {
			return fmt.Errorf("outputs are not equal")
		}
		return nil
	}
}

// ApproxEqual returns a Relation requiring both outputs to be within tol of
// each other.
func ApproxEqual[I any](tol float64) Relation[I, float64] {
	return func(c Case[I, float64]) error {
		if math.Abs(c.SourceOut-c.FollowUpOut) > tol {
			return fmt.Errorf("|%v - %v| = %v exceeds tolerance %v",
				c.SourceOut, c.FollowUpOut, math.Abs(c.SourceOut-c.FollowUpOut), tol)
		}
		return nil
	}
}

// All returns a Relation that holds when every given relation holds. The
// first violation found is reported.
func All[I, O any](rels ...Relation[I, O]) Relation[I, O] {
	return func(c Case[I, O]) error {
		for _, rel := range rels {
			if err := rel(c); err != nil {
				return err
			}
		}
		return nil
	}
}

// Any returns a Relation that holds when at least one given relation holds —
// a weaker oracle for cases where several outcomes are acceptable.
func Any[I, O any](rels ...Relation[I, O]) Relation[I, O] {
	return func(c Case[I, O]) error {
		var errs []error
		for _, rel := range rels {
			if err := rel(c); err != nil {
				errs = append(errs, err)
			} else {
				return nil
			}
		}
		return fmt.Errorf("no alternative held: %v", errs)
	}
}

// Contra returns a Relation that judges a projection of the outputs: proj is
// applied to both outputs before rel judges them. Most real relations are a
// projection plus a simple check — "same length", "same set of elements",
// "same maximum".
func Contra[I, O, P any](proj func(O) P, rel Relation[I, P]) Relation[I, O] {
	return func(c Case[I, O]) error {
		return rel(Case[I, P]{
			Source:      c.Source,
			FollowUp:    c.FollowUp,
			SourceOut:   proj(c.SourceOut),
			FollowUpOut: proj(c.FollowUpOut),
		})
	}
}

// PermutationOf returns a Relation requiring both outputs to contain the
// same elements with the same multiplicities, in any order.
func PermutationOf[I any, E comparable]() Relation[I, []E] {
	return func(c Case[I, []E]) error {
		if !sameMultiset(c.SourceOut, c.FollowUpOut) {
			return fmt.Errorf("outputs are not permutations of each other")
		}
		return nil
	}
}

// SetEqual returns a Relation requiring both outputs to contain the same set
// of elements, ignoring order and duplicates.
func SetEqual[I any, E comparable]() Relation[I, []E] {
	return func(c Case[I, []E]) error {
		src := make(map[E]bool, len(c.SourceOut))
		for _, e := range c.SourceOut {
			src[e] = true
		}
		fup := make(map[E]bool, len(c.FollowUpOut))
		for _, e := range c.FollowUpOut {
			fup[e] = true
		}
		if !maps.Equal(src, fup) {
			return fmt.Errorf("outputs do not contain the same set of elements")
		}
		return nil
	}
}

// AtLeast returns a Relation requiring the follow-up output to be greater
// than or equal to the source output: more (or bigger) input must not shrink
// the output.
func AtLeast[I any, O cmp.Ordered]() Relation[I, O] {
	return func(c Case[I, O]) error {
		if c.FollowUpOut < c.SourceOut {
			return fmt.Errorf("follow-up output %v < source output %v", c.FollowUpOut, c.SourceOut)
		}
		return nil
	}
}

// AtMost returns a Relation requiring the follow-up output to be less than
// or equal to the source output.
func AtMost[I any, O cmp.Ordered]() Relation[I, O] {
	return func(c Case[I, O]) error {
		if c.FollowUpOut > c.SourceOut {
			return fmt.Errorf("follow-up output %v > source output %v", c.FollowUpOut, c.SourceOut)
		}
		return nil
	}
}

func sameMultiset[E comparable](a, b []E) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[E]int, len(a))
	for _, e := range a {
		counts[e]++
	}
	for _, e := range b {
		counts[e]--
		if counts[e] < 0 {
			return false
		}
	}
	return true
}
