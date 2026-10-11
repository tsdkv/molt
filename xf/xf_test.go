package xf_test

import (
	"slices"
	"testing"

	"github.com/tsdkv/molt"
	"github.com/tsdkv/molt/xf"
	"pgregory.net/rapid"
)

// Exercise the transforms through real relations over the identity SUT:
// source and follow-up inputs become the compared outputs directly.

func TestShufflePreservesMultiset(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		molt.Run(rt, rapid.SliceOf(rapid.Int()), func(in []int) []int { return in },
			molt.RelCase("probe", xf.Shuffle[int](), func(c molt.Case[[]int, []int]) error {
				a, b := slices.Clone(c.FollowUp), slices.Clone(c.Source)
				slices.Sort(a)
				slices.Sort(b)
				if !slices.Equal(a, b) {
					t.Fatalf("shuffle changed the multiset: %v -> %v", c.Source, c.FollowUp)
				}
				return nil
			}))
	})
}

func TestDropRandomDropsOne(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		molt.Run(rt, rapid.SliceOf(rapid.Int()), func(in []int) []int { return in },
			molt.RelCase("probe", xf.DropRandom[int](), func(c molt.Case[[]int, []int]) error {
				want := len(c.Source)
				if want > 0 {
					want--
				}
				if len(c.FollowUp) != want {
					t.Fatalf("drop removed the wrong number of elements: %d -> %d", len(c.Source), len(c.FollowUp))
				}
				return nil
			}))
	})
}

func TestReverseIsInvolution(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		molt.Run(rt, rapid.SliceOf(rapid.Int()), func(in []int) []int { return in },
			molt.Rel("reverse-twice-is-identity",
				molt.Compose(xf.Reverse[int](), xf.Reverse[int]()),
				molt.EqualFunc[[]int](slices.Equal)))
	})
}
