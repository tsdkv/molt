package molt

// RunExamples runs each relation over a fixed list of inputs instead of
// generated ones: regression corpora harvested from rapid failures, inputs
// too structured to generate, or known-nasty cases you always want checked.
//
// Transforms receive a nil *rapid.T and therefore must not draw from it;
// relations whose transforms draw can only run under Run.
func RunExamples[I, O any](t Failer, examples []I, f func(I) O, mrs ...MR[I, O]) {
	for _, src := range examples {
		srcOut := f(src)
		for _, mr := range mrs {
			if mr.When != nil && !mr.When(src) {
				continue
			}
			fup := mr.Transform(nil, src)
			Assert(t, mr.Name, Case[I, O]{
				Source:      src,
				FollowUp:    fup,
				SourceOut:   srcOut,
				FollowUpOut: f(fup),
			}, mr.Check)
		}
	}
}

// RunExamplesN is RunExamples for n-ary relations: each example is a group
// of inputs, as with RunN.
func RunExamplesN[I, O any](t Failer, examples [][]I, f func([]I) O, mrs ...MRN[I, O]) {
	for _, srcs := range examples {
		srcOut := f(srcs)
		for _, mr := range mrs {
			if mr.When != nil && !mr.When(srcs) {
				continue
			}
			fups := mr.Transform(nil, srcs)
			AssertN(t, mr.Name, CaseN[I, O]{
				Sources:     srcs,
				FollowUps:   fups,
				SourceOut:   srcOut,
				FollowUpOut: f(fups),
			}, mr.Check)
		}
	}
}
