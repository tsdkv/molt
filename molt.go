// Package molt provides metamorphic testing for Go on top of rapid.
//
// Metamorphic testing checks how the outputs of a program relate across
// runs with different but related inputs, instead of checking outputs
// against known-correct answers. It is useful when writing down the expected
// output is hard or impossible: complex data structures, encoders, queries,
// numerical code.
//
// A metamorphic relation (MR) pairs an input transformation ("reverse the
// list") with a relation that must hold between the two resulting outputs
// ("the sorted results are identical"). Package molt lets you state such
// relations as data (see MR) and runs them against generated inputs with
// Run, or checks several implementations against each other with Diff.
//
// Because all inputs are drawn through rapid, failing cases are reproducible
// from a seed and are automatically minimized: only the source input is
// drawn, and the follow-up input is re-derived on every iteration, so
// shrinking always keeps the two consistent.
package molt

import "pgregory.net/rapid"

// Failer reports test failures. *testing.T and *rapid.T both satisfy it;
// pass a plain *testing.T to use molt outside of rapid (see RunExamples).
type Failer interface {
	Fatalf(format string, args ...any)
}

// Case is one full execution of a metamorphic relation: the source input,
// the follow-up input derived from it, and the outputs observed for both.
type Case[I, O any] struct {
	Source      I
	FollowUp    I
	SourceOut   O
	FollowUpOut O
}

// Relation judges a Case. A nil error means the relation holds.
type Relation[I, O any] func(c Case[I, O]) error

// MR is a metamorphic relation stated as data.
//
// Transform derives the follow-up input from the source input. It receives
// the rapid state so it can draw any extra randomness it needs (inserted
// elements, offsets, constants); values drawn this way are reproducible and
// shrink together with the source input. Never use math/rand here.
//
// Check decides whether the two runs satisfy the relation, returning a
// non-nil error describing the violation otherwise.
//
// When, if set, restricts the relation to source inputs satisfying the
// precondition; it is re-evaluated for every (possibly shrunk) candidate.
type MR[I, O any] struct {
	Name      string
	Transform func(t *rapid.T, in I) I
	Check     Relation[I, O]
	When      func(in I) bool
}
