package laws_test

import (
	"testing"

	"github.com/tsdkv/molt/internal/testutil"
	"github.com/tsdkv/molt/laws"
	"pgregory.net/rapid"
)

type intSet map[int]struct{}

func newIntSet(elems []int) intSet {
	s := make(intSet, len(elems))
	for _, e := range elems {
		s[e] = struct{}{}
	}
	return s
}

func intSetElements(s intSet) []int {
	out := make([]int, 0, len(s))
	for e := range s {
		out = append(out, e)
	}
	return out
}

func intSetUnion(a, b intSet) intSet {
	out := make(intSet, len(a)+len(b))
	for e := range a {
		out[e] = struct{}{}
	}
	for e := range b {
		out[e] = struct{}{}
	}
	return out
}

func intSetInter(a, b intSet) intSet {
	out := intSet{}
	for e := range a {
		if _, ok := b[e]; ok {
			out[e] = struct{}{}
		}
	}
	return out
}

func TestSet(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		laws.Set(t, rapid.IntRange(0, 20), newIntSet, intSetElements, intSetUnion, intSetInter)
	})
}

// A union that ignores its second operand must be caught.
func TestSetMutationCaught(t *testing.T) {
	brokenUnion := func(a, b intSet) intSet { return a }
	tb := &testutil.TB{}
	rapid.Check(tb, func(t *rapid.T) {
		laws.Set(t, rapid.IntRange(0, 20), newIntSet, intSetElements, brokenUnion, intSetInter)
	})
	if !tb.Failed() {
		t.Fatal("broken union was not caught")
	}
}

// Construction laws run on their own when no algebra is provided.
func TestSetConstructionOnly(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		laws.Set(t, rapid.IntRange(0, 20), newIntSet, intSetElements, nil, nil)
	})
}
