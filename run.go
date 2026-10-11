package molt

import "pgregory.net/rapid"

// Run draws a source input from gen, runs f on it, and then applies each
// metamorphic relation in turn: transforming the input, re-running f on the
// follow-up input, and checking the relation between the two runs.
//
// Only the source input is drawn; follow-up inputs are always re-derived
// from it, so rapid's minimization keeps source and follow-up consistent.
//
// Run must be called inside a rapid.Check body (or equivalent) and f must be
// deterministic: equal inputs must produce equal outputs.
func Run[I, O any](t *rapid.T, gen *rapid.Generator[I], f func(I) O, mrs ...MR[I, O]) {
	src := gen.Draw(t, "source")
	srcOut := f(src)
	for _, mr := range mrs {
		if mr.When != nil && !mr.When(src) {
			continue
		}
		fup := mr.Transform(t, src)
		Assert(t, mr.Name, Case[I, O]{
			Source:      src,
			FollowUp:    fup,
			SourceOut:   srcOut,
			FollowUpOut: f(fup),
		}, mr.Check)
	}
}
