package laws_test

import (
	"cmp"
	"testing"

	"github.com/tsdkv/molt/laws"
	"pgregory.net/rapid"
)

func TestComparator(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		laws.Comparator(t, rapid.Int(), cmp.Compare[int], func(a, b int) bool { return a == b })
	})
}

// A comparator that always reports "less" breaks reflexivity and
// antisymmetry and must be caught.
func TestComparatorMutationCaught(t *testing.T) {
	broken := func(a, b int) int { return -1 }
	tb := &fakeTB{}
	rapid.Check(tb, func(t *rapid.T) {
		laws.Comparator(t, rapid.Int(), broken, func(a, b int) bool { return a == b })
	})
	if !tb.Failed() {
		t.Fatal("broken comparator was not caught")
	}
}
