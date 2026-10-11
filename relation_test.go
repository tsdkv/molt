package molt_test

import (
	"errors"
	"testing"

	"github.com/tsdkv/molt"
)

func TestRelationPack(t *testing.T) {
	eq := molt.Equal[int]()
	if err := eq(1, 1); err != nil {
		t.Fatal(err)
	}
	if err := eq(1, 2); err == nil {
		t.Fatal("Equal accepted different values")
	}
	approx := molt.ApproxEqual(0.5)
	if err := approx(1.0, 1.4); err != nil {
		t.Fatal(err)
	}
	if err := approx(1.0, 2.0); err == nil {
		t.Fatal("ApproxEqual accepted values beyond tolerance")
	}
	if err := molt.AtLeast[int]()(1, 1); err != nil {
		t.Fatal(err)
	}
	if err := molt.AtLeast[int]()(2, 1); err == nil {
		t.Fatal("AtLeast accepted a smaller follow-up")
	}
	if err := molt.AtMost[int]()(1, 2); err == nil {
		t.Fatal("AtMost accepted a larger follow-up")
	}
	if err := molt.PermutationOf[int]()([]int{1, 2, 2}, []int{2, 1, 2}); err != nil {
		t.Fatal(err)
	}
	if err := molt.PermutationOf[int]()([]int{1, 2, 2}, []int{1, 1, 2}); err == nil {
		t.Fatal("PermutationOf accepted different multiplicities")
	}
	if err := molt.SetEqual[int]()([]int{1, 2, 2}, []int{2, 1}); err != nil {
		t.Fatal(err)
	}
	if err := molt.SetEqual[int]()([]int{1, 2}, []int{1, 3}); err == nil {
		t.Fatal("SetEqual accepted different sets")
	}
}

func TestRelationCombinators(t *testing.T) {
	ok := func(a, b int) error { return nil }
	bad := func(a, b int) error { return errors.New("nope") }

	if err := molt.All(ok, ok)(1, 2); err != nil {
		t.Fatal(err)
	}
	if err := molt.All(ok, bad)(1, 2); err == nil {
		t.Fatal("All passed with a failing relation")
	}
	if err := molt.Any(ok, bad)(1, 2); err != nil {
		t.Fatal(err)
	}
	if err := molt.Any(bad, bad)(1, 2); err == nil {
		t.Fatal("Any passed with all relations failing")
	}

	double := molt.Contra(func(x int) int { return x * 2 }, molt.Equal[int]())
	if err := double(3, 3); err != nil {
		t.Fatal(err)
	}
	if err := double(3, 4); err == nil {
		t.Fatal("Contra passed on different projections")
	}
}
