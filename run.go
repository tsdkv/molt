package molt

import (
	"testing"

	"pgregory.net/rapid"
)

// Test runs each relation as its own subtest with its own rapid session:
// independent seeds, independent minimization, and failure attribution via
// `go test -run TestName/relation-name`. It is the default entry point.
//
// f must be deterministic: equal inputs must produce equal outputs.
func Test[I, O any](tb *testing.T, gen *rapid.Generator[I], f func(I) O, mrs ...MR[I, O]) {
	tb.Helper()
	for _, mr := range mrs {
		tb.Run(mr.name, func(tb *testing.T) {
			var total, skipped int
			rapid.Check(tb, func(t *rapid.T) {
				src := gen.Draw(t, "source")
				total++
				if mr.when != nil && !mr.when(src) {
					skipped++
					return
				}
				checkOne(t, mr, f, src)
			})
			if skipped*2 > total {
				tb.Logf("molt: precondition of %q filtered out %d/%d inputs; consider a tighter generator",
					mr.name, skipped, total)
			}
		})
	}
}

// Run draws a source input from gen inside the current rapid.Check body and
// applies each relation in turn. Use it in free-form rapid tests that do
// not fit the Test shape: several SUTs, mixed draws, custom control flow.
//
// Only the source input is drawn; follow-up inputs are always re-derived
// from it, so rapid's minimization keeps source and follow-up consistent.
func Run[I, O any](t *rapid.T, gen *rapid.Generator[I], f func(I) O, mrs ...MR[I, O]) {
	src := gen.Draw(t, "source")
	for _, mr := range mrs {
		if mr.when != nil && !mr.when(src) {
			continue
		}
		checkOne(t, mr, f, src)
	}
}

// RunExamples runs each relation over a fixed list of inputs instead of
// generated ones: regression corpora harvested from rapid failures, inputs
// too structured to generate, or known-nasty cases. It needs no rapid at
// all — a plain *testing.T works. Relations using drawing transforms are
// rejected up front.
func RunExamples[I, O any](tb interface {
	Helper()
	Fatalf(format string, args ...any)
}, examples []I, f func(I) O, mrs ...MR[I, O]) {
	tb.Helper()
	for _, mr := range mrs {
		if mr.tx.draws {
			tb.Fatalf("molt: relation %q uses a drawing transform and cannot run on fixed examples", mr.name)
			return
		}
	}
	for _, src := range examples {
		for _, mr := range mrs {
			if mr.when != nil && !mr.when(src) {
				continue
			}
			fup := mr.tx.followUp(nil, src)
			srcOut, fupOut := f(src), f(fup)
			if err := mr.check(Case[I, O]{src, fup, srcOut, fupOut}); err != nil {
				tb.Fatalf("metamorphic relation %q violated on example: %v\n"+
					"source:        %#v\nfollow-up:     %#v\nsource out:    %#v\nfollow-up out: %#v",
					mr.name, err, src, fup, srcOut, fupOut)
			}
		}
	}
}

func checkOne[I, O any](t *rapid.T, mr MR[I, O], f func(I) O, src I) {
	fup := mr.tx.followUp(t, src)
	srcOut, fupOut := f(src), f(fup)
	if err := mr.check(Case[I, O]{src, fup, srcOut, fupOut}); err != nil {
		t.Fatalf("metamorphic relation %q violated: %v\n"+
			"source:        %#v\nfollow-up:     %#v\nsource out:    %#v\nfollow-up out: %#v",
			mr.name, err, src, fup, srcOut, fupOut)
	}
}
