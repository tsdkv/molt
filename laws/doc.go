// Package laws provides ready-made metamorphic relation sets for contracts
// that Go code re-implements all the time: codecs, comparators, sort
// functions, key-value stores.
//
// Each function runs a battery of relations against the provided
// implementation via molt: every law is checked on generated inputs, and
// violations are minimized by rapid. A law set is only shipped once it has
// been shown to catch deliberately broken implementations — see the package
// tests.
package laws
