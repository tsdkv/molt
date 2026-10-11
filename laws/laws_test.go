package laws_test

import (
	"cmp"
	"encoding/json"
	"maps"
	"slices"
	"testing"

	"github.com/tsdkv/molt/laws"
	"pgregory.net/rapid"
)

// --- Codec ---

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

func marshalDoc(d doc) ([]byte, error) { return json.Marshal(d) }

func unmarshalDoc(b []byte) (doc, error) {
	var d doc
	err := json.Unmarshal(b, &d)
	return d, err
}

func TestCodec(t *testing.T) {
	laws.Codec(t, docGen(), marshalDoc, unmarshalDoc, func(a, b doc) bool { return a == b })
}

// --- Sortable ---

func TestSortable(t *testing.T) {
	laws.Sortable(t, rapid.SliceOf(rapid.Int()), func(in []int) []int {
		out := slices.Clone(in)
		slices.Sort(out)
		return out
	})
}

// --- Comparator ---

func TestComparator(t *testing.T) {
	laws.Comparator(t, rapid.Int(), cmp.Compare[int],
		func(a, b int) bool { return a == b })
}

// --- Set ---

type intSet map[int]bool

func newIntSet(elems []int) intSet {
	s := intSet{}
	for _, e := range elems {
		s[e] = true
	}
	return s
}

func intSetElements(s intSet) []int { return slices.Collect(maps.Keys(s)) }

func intSetUnion(a, b intSet) intSet {
	out := intSet{}
	for e := range a {
		out[e] = true
	}
	for e := range b {
		out[e] = true
	}
	return out
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

func TestSet(t *testing.T) {
	laws.Set(t, rapid.IntRange(0, 100), newIntSet, intSetElements, intSetUnion, intSetInter)
}

func TestSetConstructionOnly(t *testing.T) {
	laws.Set(t, rapid.IntRange(0, 100), newIntSet, intSetElements, nil, nil)
}
