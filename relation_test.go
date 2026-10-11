package molt_test

import (
	"errors"
	"testing"

	"github.com/tsdkv/molt"
)

func TestRelationCombinators(t *testing.T) {
	c := molt.Case[int, []int]{Source: 1, FollowUp: 2, SourceOut: []int{1, 2, 3}, FollowUpOut: []int{3, 2, 1}}

	lengths := molt.Contra(func(out []int) int { return len(out) }, molt.Equal[int, int]())
	if err := lengths(c); err != nil {
		t.Fatalf("Contra(len, Equal) should hold: %v", err)
	}

	passing := molt.Equal[int, int]()
	failing := func(molt.Case[int, int]) error { return errors.New("nope") }

	if err := molt.All(passing, failing)(molt.Case[int, int]{}); err == nil {
		t.Fatal("All should fail when any relation fails")
	}
	if err := molt.Any(failing, passing)(molt.Case[int, int]{}); err != nil {
		t.Fatalf("Any should hold when one relation holds: %v", err)
	}
	if err := molt.Any(failing, failing)(molt.Case[int, int]{}); err == nil {
		t.Fatal("Any should fail when every relation fails")
	}
}

func TestRelationPack(t *testing.T) {
	c := molt.Case[int, []int]{SourceOut: []int{3, 1, 2}, FollowUpOut: []int{2, 3, 1}}
	if err := molt.PermutationOf[int, int]()(c); err != nil {
		t.Fatalf("PermutationOf should hold: %v", err)
	}
	dup := molt.Case[int, []int]{SourceOut: []int{1, 1, 2}, FollowUpOut: []int{1, 2, 2}}
	if err := molt.PermutationOf[int, int]()(dup); err == nil {
		t.Fatal("PermutationOf should fail when multiplicities differ")
	}
	if err := molt.SetEqual[int, int]()(dup); err != nil {
		t.Fatalf("SetEqual should ignore multiplicities: %v", err)
	}
	shuffled := molt.Case[int, []int]{SourceOut: []int{1, 2}, FollowUpOut: []int{2, 1}}
	if err := molt.SetEqual[int, int]()(shuffled); err != nil {
		t.Fatalf("SetEqual should ignore order: %v", err)
	}

	up := molt.Case[int, int]{SourceOut: 3, FollowUpOut: 5}
	if err := molt.AtLeast[int, int]()(up); err != nil {
		t.Fatalf("AtLeast should hold: %v", err)
	}
	if err := molt.AtMost[int, int]()(up); err == nil {
		t.Fatal("AtMost should fail")
	}
}
