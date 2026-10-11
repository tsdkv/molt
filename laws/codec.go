package laws

import (
	"bytes"
	"fmt"

	"github.com/tsdkv/molt"
	"pgregory.net/rapid"
)

// Codec checks the standard contract for a marshal/unmarshal pair: encoding
// is deterministic, and decoding an encoding returns the original value.
// marshal must succeed for every value gen can produce.
func Codec[I any](t *rapid.T, gen *rapid.Generator[I], marshal func(I) ([]byte, error), unmarshal func([]byte) (I, error), eq func(a, b I) bool) {
	f := func(in I) []byte {
		b, err := marshal(in)
		if err != nil {
			panic(fmt.Errorf("laws.Codec: marshal(%#v) failed: %w", in, err))
		}
		return b
	}
	molt.Run(t, gen, f,
		molt.MR[I, []byte]{
			Name:      "encode-deterministic",
			Transform: molt.Identity[I],
			Check:     molt.EqualFunc[I, []byte](bytes.Equal),
		},
		molt.MR[I, []byte]{
			Name:      "round-trip",
			Transform: molt.Identity[I],
			Check: func(c molt.Case[I, []byte]) error {
				back, err := unmarshal(c.SourceOut)
				if err != nil {
					return fmt.Errorf("unmarshal of own encoding failed: %w", err)
				}
				if !eq(c.Source, back) {
					return fmt.Errorf("round-trip changed the value: %#v -> %#v", c.Source, back)
				}
				return nil
			},
		},
	)
}
