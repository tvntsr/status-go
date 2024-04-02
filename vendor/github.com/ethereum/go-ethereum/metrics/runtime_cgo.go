//go:build cgo && !appengine && !wasm
// +build cgo,!appengine,!wasm

package metrics

import "runtime"

func numCgoCall() int64 {
	return runtime.NumCgoCall()
}
