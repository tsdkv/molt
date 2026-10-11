package molt_test

import (
	"testing"

	"github.com/tsdkv/molt"
	"pgregory.net/rapid"
)

type kvStore interface {
	Set(k string, v int)
	Get(k string) (int, bool)
	Del(k string)
}

type mapStore struct{ m map[string]int }

func newMapStore() *mapStore            { return &mapStore{m: map[string]int{}} }
func (s *mapStore) Set(k string, v int) { s.m[k] = v }
func (s *mapStore) Get(k string) (int, bool) {
	v, ok := s.m[k]
	return v, ok
}
func (s *mapStore) Del(k string) { delete(s.m, k) }

// sliceStore is a deliberately unoptimized but correct implementation.
type sliceStore struct {
	keys []string
	vals []int
}

func (s *sliceStore) Set(k string, v int) {
	for i, key := range s.keys {
		if key == k {
			s.vals[i] = v
			return
		}
	}
	s.keys = append(s.keys, k)
	s.vals = append(s.vals, v)
}

func (s *sliceStore) Get(k string) (int, bool) {
	for i, key := range s.keys {
		if key == k {
			return s.vals[i], true
		}
	}
	return 0, false
}

func (s *sliceStore) Del(k string) {
	for i, key := range s.keys {
		if key == k {
			s.keys = append(s.keys[:i], s.keys[i+1:]...)
			s.vals = append(s.vals[:i], s.vals[i+1:]...)
			return
		}
	}
}

type getResult struct {
	v  int
	ok bool
}

func kvOps[S kvStore]() []molt.Op[S] {
	key := func(t *rapid.T) string {
		return string(rune('a' + rapid.IntRange(0, 4).Draw(t, "key")))
	}
	return []molt.Op[S]{
		{
			Name: "set", Weight: 3,
			Prepare: func(t *rapid.T) func(S) any {
				k, v := key(t), rapid.Int().Draw(t, "val")
				return func(s S) any { s.Set(k, v); return nil }
			},
		},
		{
			Name: "get", Weight: 3,
			Prepare: func(t *rapid.T) func(S) any {
				k := key(t)
				return func(s S) any {
					v, ok := s.Get(k)
					return getResult{v, ok}
				}
			},
		},
		{
			Name: "del", Weight: 1,
			Prepare: func(t *rapid.T) func(S) any {
				k := key(t)
				return func(s S) any { s.Del(k); return nil }
			},
		},
	}
}

func TestKVDifferential(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		molt.Diff(t,
			[]molt.Impl[kvStore]{
				{Name: "map", Open: func(t *rapid.T) kvStore { return newMapStore() }},
				{Name: "slice", Open: func(t *rapid.T) kvStore { return &sliceStore{} }},
			},
			kvOps[kvStore](),
			rapid.IntRange(1, 100),
		)
	})
}

// A store whose Del is a no-op must diverge from the correct model.
func TestKVMutationCaught(t *testing.T) {
	tb := &fakeTB{}
	rapid.Check(tb, func(t *rapid.T) {
		molt.Diff(t,
			[]molt.Impl[kvStore]{
				{Name: "map", Open: func(t *rapid.T) kvStore { return newMapStore() }},
				{Name: "broken-slice", Open: func(t *rapid.T) kvStore { return &brokenDelStore{} }},
			},
			kvOps[kvStore](),
			rapid.IntRange(1, 100),
		)
	})
	if !tb.Failed() {
		t.Fatal("store with broken Del was not caught")
	}
}

type brokenDelStore struct{ sliceStore }

func (s *brokenDelStore) Del(string) {}
