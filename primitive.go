package molt

import "pgregory.net/rapid"

// Follow draws a source input from gen and derives the follow-up input from
// it with transform. Only the source is drawn; the follow-up is re-derived
// on every rapid iteration, so shrinking always keeps the two consistent.
//
// Follow is the primitive Run is built on; use it directly when writing
// free-form rapid tests that do not fit the Run shape.
func Follow[I any](t *rapid.T, gen *rapid.Generator[I], transform func(t *rapid.T, in I) I) (src, fup I) {
	src = gen.Draw(t, "source")
	fup = transform(t, src)
	return src, fup
}

// Execute runs f on the source and follow-up inputs and bundles all four
// artifacts of a metamorphic execution into a Case.
func Execute[I, O any](f func(I) O, src, fup I) Case[I, O] {
	return Case[I, O]{
		Source:      src,
		FollowUp:    fup,
		SourceOut:   f(src),
		FollowUpOut: f(fup),
	}
}

// Assert judges a Case with rel and fails the test — reporting the relation
// name and all four artifacts — if the relation does not hold.
func Assert[I, O any](t *rapid.T, name string, c Case[I, O], rel Relation[I, O]) {
	if err := rel(c); err != nil {
		t.Fatalf("metamorphic relation %q violated: %v\n"+
			"source:         %#v\n"+
			"follow-up:      %#v\n"+
			"source out:     %#v\n"+
			"follow-up out:  %#v",
			name, err, c.Source, c.FollowUp, c.SourceOut, c.FollowUpOut)
	}
}
