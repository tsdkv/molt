package laws_test

import (
	"encoding/json"
	"testing"

	"github.com/tsdkv/molt/laws"
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

func marshalDoc(d doc) ([]byte, error) { return json.Marshal(d) }

func unmarshalDoc(b []byte) (doc, error) {
	var d doc
	err := json.Unmarshal(b, &d)
	return d, err
}

func TestCodec(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		laws.Codec(t, docGen(), marshalDoc, unmarshalDoc, func(a, b doc) bool { return a == b })
	})
}

// A decoder that corrupts the value must be caught by the round-trip law.
func TestCodecMutationCaught(t *testing.T) {
	broken := func(b []byte) (doc, error) {
		d, err := unmarshalDoc(b)
		d.N++
		return d, err
	}
	tb := &fakeTB{}
	rapid.Check(tb, func(t *rapid.T) {
		laws.Codec(t, docGen(), marshalDoc, broken, func(a, b doc) bool { return a == b })
	})
	if !tb.Failed() {
		t.Fatal("corrupting decoder was not caught")
	}
}
