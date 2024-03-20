// +build js
// ::FIXME:: it is not available for js
package mmap

import (
	"fmt"
)


func mmap(len int, prot, flags, hfile uintptr, off int64) ([]byte, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (m MMap) flush() error {
	return nil
}

func (m MMap) lock() error {
	return nil
}

func (m MMap) unlock() error {
	return nil
}

func (m MMap) unmap() error {
	return nil
}
