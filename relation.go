package molt

import (
	"fmt"
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
