package laws

// White-box mutation tests: deliberately broken implementations must be
// caught by the relation batteries. Failures are recorded with testutil.TB
// instead of aborting the test.

import (
	"encoding/json"
	"maps"
	"slices"
	"testing"

	"github.com/tsdkv/molt"
	"github.com/tsdkv/molt/internal/testutil"
	"pgregory.net/rapid"
)

type doc struct {
	Title string
	N     int
}

func docGen() *rapid.Generator[doc] {
	return rapid.Custom(func(t *rapid.T) doc {
		return doc{
			Title: rapid.StringN(0, 10, -1).Draw(t, "title"),
			N:     rapid.Int().Draw(t, "n"),
		}
	})
}

func TestCodecMutationCaught(t *testing.T) {
	broken := func(b []byte) (doc, error) {
		var d doc
		err := json.Unmarshal(b, &d)
		d.N++
		return d, err
	}
	marshal := func(d doc) ([]byte, error) { return json.Marshal(d) }
	f := func(in doc) []byte {
		b, err := marshal(in)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	tb := &testutil.TB{}
	rapid.Check(tb, func(t *rapid.T) {
		molt.Run(t, docGen(), f, codecMRs(marshal, broken, func(a, b doc) bool { return a == b })...)
	})
	if !tb.Failed() {
		t.Fatal("corrupting decoder was not caught")
	}
}

func TestSortableMutationCaught(t *testing.T) {
	broken := func(in []int) []int {
		out := slices.Clone(in)
		slices.Sort(out)
		if len(out) > 1 {
			out = out[1:] // drops the minimum
		}
		return out
	}
	tb := &testutil.TB{}
	rapid.Check(tb, func(t *rapid.T) {
		molt.Run(t, rapid.SliceOf(rapid.Int()), broken, sortableMRs(broken)...)
	})
	if !tb.Failed() {
		t.Fatal("broken sort was not caught")
	}
}

func TestComparatorMutationCaught(t *testing.T) {
	broken := func(a, b int) int { return 1 } // never reflexive, never consistent
	tb := &testutil.TB{}
	rapid.Check(tb, func(t *rapid.T) {
		molt.Run(t, rapid.Int(), func(x int) int { return broken(x, x) }, reflexiveMR[int]())
		molt.Run(t, rapid.SliceOfN(rapid.Int(), 2, 2),
			func(xs []int) int { return broken(xs[0], xs[1]) },
			pairMRs(func(a, b int) bool { return a == b })...)
	})
	if !tb.Failed() {
		t.Fatal("broken comparator was not caught")
	}
}

type intSet map[int]bool

func newIntSet(elems []int) intSet {
	s := intSet{}
	for _, e := range elems {
		s[e] = true
	}
	return s
}

func intSetInter(a, b intSet) intSet {
	out := intSet{}
	for e := range a {
		if b[e] {
			out[e] = true
		}
	}
	return out
}

func canonIntSet(s intSet) []int {
	out := slices.Collect(maps.Keys(s))
	slices.Sort(out)
	return out
}

func TestSetMutationCaught(t *testing.T) {
	brokenUnion := func(a, b intSet) intSet { return a } // forgets the second operand
	canonList := func(elems []int) []int { return canonIntSet(newIntSet(elems)) }
	tb := &testutil.TB{}
	rapid.Check(tb, func(t *rapid.T) {
		molt.Run(t, rapid.SliceOfN(rapid.SliceOf(rapid.IntRange(0, 100)), 2, 2),
			func(srcs [][]int) setPair[int] {
				a, b := newIntSet(srcs[0]), newIntSet(srcs[1])
				return setPair[int]{canonIntSet(brokenUnion(a, b)), canonIntSet(intSetInter(a, b))}
			},
			algebraMRs(brokenUnion, intSetInter, canonList)...)
	})
	if !tb.Failed() {
		t.Fatal("broken union was not caught")
	}
}
