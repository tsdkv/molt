// Package laws ships ready-made batteries of metamorphic relations for
// contracts Go code re-implements all the time: codecs, sort functions,
// comparators, sets. Each entry point is one line in a test and needs no
// rapid knowledge:
//
//	func TestMySort(t *testing.T) {
//		laws.Sortable(t, rapid.SliceOf(rapid.Int()), mysort)
//	}
package laws
