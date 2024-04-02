// +build wasm

package fileutil

import (
	"os"
)

func preallocExtend(f *os.File, sizeInBytes int64) error {
	return nil
}

func preallocFixed(f *os.File, sizeInBytes int64) error {
	return nil
}
