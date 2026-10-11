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
