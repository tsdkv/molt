package molt

import "pgregory.net/rapid"

// CaseN is one execution of an n-ary metamorphic relation: k independent
// source inputs, the follow-up inputs derived from them, and the outputs the
// system produced for each group.
type CaseN[I, O any] struct {
	Sources     []I
	FollowUps   []I
	SourceOut   O
	FollowUpOut O
}

// RelationN judges a CaseN. A nil error means the relation holds.
type RelationN[I, O any] func(c CaseN[I, O]) error

// MRN is an n-ary metamorphic relation stated as data: the system is a
// function of k inputs, Transform derives the follow-up input group from the
// sources (it may change the group size, reorder, or substitute elements),
// and Check relates the two runs.
//
// This covers relations that need several independent inputs — commutativity
// (swap the group), transitivity (three inputs, one check) — which a
// single-source MR cannot express.
type MRN[I, O any] struct {
	Name      string
	Transform func(t *rapid.T, srcs []I) []I
	Check     RelationN[I, O]
	When      func(srcs []I) bool
}

// FollowN draws k independent source inputs and derives the follow-up group
// from them. Only the sources are drawn; follow-ups are re-derived on every
// rapid iteration, so shrinking keeps the groups consistent.
func FollowN[I any](t *rapid.T, gen *rapid.Generator[I], k int, transform func(t *rapid.T, srcs []I) []I) (srcs, fups []I) {
	srcs = make([]I, k)
	for i := range srcs {
		srcs[i] = gen.Draw(t, "source")
	}
	fups = transform(t, srcs)
	return srcs, fups
}

// ExecuteN runs f on the source group and the follow-up group and bundles
// the artifacts into a CaseN.
func ExecuteN[I, O any](f func([]I) O, srcs, fups []I) CaseN[I, O] {
	return CaseN[I, O]{
		Sources:     srcs,
		FollowUps:   fups,
		SourceOut:   f(srcs),
		FollowUpOut: f(fups),
	}
}

// AssertN judges a CaseN with rel and fails the test — reporting the
// relation name and all artifacts — if the relation does not hold.
func AssertN[I, O any](t Failer, name string, c CaseN[I, O], rel RelationN[I, O]) {
	if err := rel(c); err != nil {
		t.Fatalf("metamorphic relation %q violated: %v\n"+
			"sources:        %#v\n"+
			"follow-ups:     %#v\n"+
			"source out:     %#v\n"+
			"follow-up out:  %#v",
			name, err, c.Sources, c.FollowUps, c.SourceOut, c.FollowUpOut)
	}
}

// RunN draws k independent source inputs per iteration, runs f on the group,
// and applies each n-ary relation in turn. Each source is drawn
// independently and shrinks independently; the follow-up group is always
// re-derived.
func RunN[I, O any](t *rapid.T, gen *rapid.Generator[I], k int, f func([]I) O, mrs ...MRN[I, O]) {
	srcs := make([]I, k)
	for i := range srcs {
		srcs[i] = gen.Draw(t, "source")
	}
	srcOut := f(srcs)
	for _, mr := range mrs {
		if mr.When != nil && !mr.When(srcs) {
			continue
		}
		fups := mr.Transform(t, srcs)
		AssertN(t, mr.Name, CaseN[I, O]{
			Sources:     srcs,
			FollowUps:   fups,
			SourceOut:   srcOut,
			FollowUpOut: f(fups),
		}, mr.Check)
	}
}
