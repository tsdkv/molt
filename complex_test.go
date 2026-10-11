package molt_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/tsdkv/molt"
	"github.com/tsdkv/molt/xf"
	"pgregory.net/rapid"
)

// A complex config as the input: the generator is plain rapid, transforms
// reach into fields with molt.On.
type config struct {
	Host    string
	Port    int
	Tags    []string
	Timeout int
}

var configGen = rapid.Custom(func(t *rapid.T) config {
	return config{
		Host:    rapid.StringMatching(`[a-z]{1,8}`).Draw(t, "host"),
		Port:    rapid.IntRange(1, 65535).Draw(t, "port"),
		Tags:    rapid.SliceOf(rapid.StringMatching(`[a-z]{1,4}`)).Draw(t, "tags"),
		Timeout: rapid.IntRange(0, 3600).Draw(t, "timeout"),
	}
})

// buildRoutes is the SUT: the sorted list of "host:port/tag" routes.
func buildRoutes(cfg config) []string {
	routes := make([]string, len(cfg.Tags))
	for i, tag := range cfg.Tags {
		routes[i] = fmt.Sprintf("%s:%d/%s", cfg.Host, cfg.Port, tag)
	}
	slices.Sort(routes)
	return routes
}

func TestBuildRoutes(t *testing.T) {
	molt.Test(t, configGen, buildRoutes,
		molt.Rel("tag-order-invariant", molt.On(xf.Shuffle[string](),
			func(c config) []string { return c.Tags },
			func(c config, tags []string) config { c.Tags = tags; return c },
		), molt.EqualFunc[[]string](slices.Equal)),
		molt.Rel("tag-dup-only-adds", molt.On(xf.Duplicate[string](),
			func(c config) []string { return c.Tags },
			func(c config, tags []string) config { c.Tags = tags; return c },
		), func(srcOut, fupOut []string) error {
			for _, r := range srcOut {
				if !slices.Contains(fupOut, r) {
					return fmt.Errorf("route %q lost after duplicating tags", r)
				}
			}
			return nil
		}),
		molt.Rel("timeout-irrelevant", molt.Pure(func(c config) config {
			c.Timeout++
			return c
		}), molt.EqualFunc[[]string](slices.Equal)),
	)
}

// An op-sequence workload as the input; complex transform randomness is a
// struct drawn with rapid.Custom.
type op struct {
	push bool // false = pop
	val  int
}

var opGen = rapid.Custom(func(t *rapid.T) op {
	if rapid.Bool().Draw(t, "push?") {
		return op{push: true, val: rapid.Int().Draw(t, "val")}
	}
	return op{}
})

// runStack is the SUT: run the ops against a stack (pop on empty is a
// no-op), return the final contents.
func runStack(ops []op) []int {
	var stack []int
	for _, o := range ops {
		if o.push {
			stack = append(stack, o.val)
		} else if len(stack) > 0 {
			stack = stack[:len(stack)-1]
		}
	}
	return stack
}

type insertion struct {
	pos uint64 // normalized modulo len inside apply
	val int
}

var insertionGen = rapid.Custom(func(t *rapid.T) insertion {
	return insertion{
		pos: rapid.Uint64().Draw(t, "pos"),
		val: rapid.Int().Draw(t, "val"),
	}
})

func TestRunStack(t *testing.T) {
	molt.Test(t, rapid.SliceOf(opGen), runStack,
		// inserting a push immediately followed by a pop, anywhere in the
		// sequence, must leave the final stack unchanged
		molt.Rel("push-pop-pair-invariant", molt.Drawn(insertionGen,
			func(in []op, r insertion) []op {
				pos := int(r.pos % uint64(len(in)+1))
				out := slices.Clone(in)
				out = slices.Insert(out, pos, op{push: true, val: r.val})
				return slices.Insert(out, pos+1, op{})
			}), molt.EqualFunc[[]int](slices.Equal)),
		// appending only pops can only shrink the stack, prefix-wise
		molt.Rel("trailing-pops-only-shrink", molt.Drawn(rapid.IntRange(1, 8),
			func(in []op, n int) []op {
				out := slices.Clone(in)
				for range n {
					out = append(out, op{})
				}
				return out
			}), func(srcOut, fupOut []int) error {
			if len(fupOut) > len(srcOut) {
				return fmt.Errorf("stack grew after trailing pops: %d -> %d", len(srcOut), len(fupOut))
			}
			if !slices.Equal(srcOut[:len(fupOut)], fupOut) {
				return fmt.Errorf("trailing pops changed surviving elements")
			}
			return nil
		}),
	)
}
