//go:build !cgo || appengine || wasm
// +build !cgo appengine wasm

package metrics

func numCgoCall() int64 {
	return 0
}
