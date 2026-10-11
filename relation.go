package molt

import (
	"cmp"
	"fmt"
	"maps"
	"math"
)

// Relation judges two outputs — source output and follow-up output. A nil
// error means the relation holds. Relations that need the inputs are built
// with RelCase and receive the full Case instead.
type Relation[O any] func(srcOut, fupOut O) error

// Equal returns a Relation requiring both outputs to be ==.
func Equal[O comparable]() Relation[O] {
	return func(a, b O) error {
		if a != b {
			return fmt.Errorf("outputs are not equal")
		}
		return nil
	}
}

// EqualFunc returns a Relation requiring both outputs to be equal according
// to eq (for example slices.Equal).
func EqualFunc[O any](eq func(a, b O) bool) Relation[O] {
	return func(a, b O) error {
		if !eq(a, b) {
			return fmt.Errorf("outputs are not equal")
		}
		return nil
	}
}

// ApproxEqual returns a Relation requiring both outputs to be within tol of
// each other.
func ApproxEqual(tol float64) Relation[float64] {
	return func(a, b float64) error {
		if math.Abs(a-b) > tol {
			return fmt.Errorf("|%v - %v| = %v exceeds tolerance %v", a, b, math.Abs(a-b), tol)
		}
		return nil
	}
}

// AtLeast returns a Relation requiring the follow-up output to be greater
// than or equal to the source output: more (or bigger) input must not
// shrink the output.
func AtLeast[O cmp.Ordered]() Relation[O] {
	return func(a, b O) error {
		if b < a {
			return fmt.Errorf("follow-up output %v < source output %v", b, a)
		}
		return nil
	}
}

// AtMost returns a Relation requiring the follow-up output to be less than
// or equal to the source output.
func AtMost[O cmp.Ordered]() Relation[O] {
	return func(a, b O) error {
		if b > a {
			return fmt.Errorf("follow-up output %v > source output %v", b, a)
		}
		return nil
	}
}

// PermutationOf returns a Relation requiring both outputs to contain the
// same elements with the same multiplicities, in any order.
func PermutationOf[E comparable]() Relation[[]E] {
	return func(a, b []E) error {
		if len(a) != len(b) {
			return fmt.Errorf("outputs are not permutations of each other")
		}
		counts := make(map[E]int, len(a))
		for _, e := range a {
			counts[e]++
		}
		for _, e := range b {
			counts[e]--
			if counts[e] < 0 {
				return fmt.Errorf("outputs are not permutations of each other")
			}
		}
		return nil
	}
}

// SetEqual returns a Relation requiring both outputs to contain the same
// set of elements, ignoring order and duplicates.
func SetEqual[E comparable]() Relation[[]E] {
	return func(a, b []E) error {
		as := make(map[E]bool, len(a))
		for _, e := range a {
			as[e] = true
		}
		bs := make(map[E]bool, len(b))
		for _, e := range b {
			bs[e] = true
		}
		if !maps.Equal(as, bs) {
			return fmt.Errorf("outputs do not contain the same set of elements")
		}
		return nil
	}
}

// All returns a Relation that holds when every given relation holds. The
// first violation found is reported.
func All[O any](rels ...Relation[O]) Relation[O] {
	return func(a, b O) error {
		for _, rel := range rels {
			if err := rel(a, b); err != nil {
				return err
			}
		}
		return nil
	}
}

// Any returns a Relation that holds when at least one given relation
// holds — a weaker oracle for cases where several outcomes are acceptable.
func Any[O any](rels ...Relation[O]) Relation[O] {
	return func(a, b O) error {
		var errs []error
		for _, rel := range rels {
			if err := rel(a, b); err != nil {
				errs = append(errs, err)
			} else {
				return nil
			}
		}
		return fmt.Errorf("no alternative held: %v", errs)
	}
}

// Contra returns a Relation that judges a projection of the outputs: proj
// is applied to both outputs before rel judges them. Most real relations
// are a projection plus a simple check — "same length", "same set of
// elements", "same maximum".
func Contra[O, P any](proj func(O) P, rel Relation[P]) Relation[O] {
	return func(a, b O) error {
		return rel(proj(a), proj(b))
	}
}
