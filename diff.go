package molt

import (
	"reflect"

	"pgregory.net/rapid"
)

// Impl is one implementation under differential test, identified by name
// ("inmem", "ondisk", ...). Open must return a fresh, empty instance; use
// t.Cleanup inside Open to release resources.
type Impl[S any] struct {
	Name string
	Open func(t *rapid.T) S
}

// Op is one kind of operation applied to every implementation.
//
// Prepare draws the operation's arguments from t exactly once and returns a
// closure applying the operation to a single system, returning an observable
// result that is compared across implementations. Return nil for operations
// with no observable output.
type Op[S any] struct {
	Name    string
	Weight  int
	Prepare func(t *rapid.T) func(s S) any
}

// Diff runs a random sequence of operations against every implementation in
// lockstep and fails at the first step where their observable results
// diverge. At least two implementations are required, and every Op must
// have a positive weight. Results are compared with equal; pass nil to use
// reflect.DeepEqual.
//
// The whole sequence — its length, the choice of operations and their
// arguments — is drawn through rapid, so divergences are reproducible from
// a seed and minimized automatically.
func Diff[S any](t *rapid.T, impls []Impl[S], ops []Op[S], steps *rapid.Generator[int], equal func(a, b any) bool) {
	if equal == nil {
		equal = func(a, b any) bool { return reflect.DeepEqual(a, b) }
	}
	if len(impls) < 2 {
		t.Fatalf("molt.Diff: need at least 2 implementations, got %d", len(impls))
	}
	total := 0
	for _, op := range ops {
		if op.Weight <= 0 {
			t.Fatalf("molt.Diff: op %q has non-positive weight %d", op.Name, op.Weight)
		}
		total += op.Weight
	}

	systems := make([]S, len(impls))
	for i, impl := range impls {
		systems[i] = impl.Open(t)
	}

	n := steps.Draw(t, "steps")
	for i := range n {
		op := pickOp(t, ops, total)
		apply := op.Prepare(t)
		results := make([]any, len(systems))
		for j, s := range systems {
			results[j] = apply(s)
		}
		for j := 1; j < len(results); j++ {
			if !equal(results[0], results[j]) {
				t.Fatalf("step %d: op %q diverged:\n"+
					"%s: %#v\n"+
					"%s: %#v",
					i, op.Name, impls[0].Name, results[0], impls[j].Name, results[j])
			}
		}
	}
}

func pickOp[S any](t *rapid.T, ops []Op[S], total int) Op[S] {
	w := rapid.IntRange(1, total).Draw(t, "op")
	for _, op := range ops {
		if w <= op.Weight {
			return op
		}
		w -= op.Weight
	}
	panic("unreachable")
}
