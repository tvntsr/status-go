// +build js

package fileutil

import (
	"os"
	"fmt"
)

func mmap(f *os.File, size int) ([]byte, error) {
	return nil, fmt.Errorf("Not implemented")
}

func munmap(b []byte) error {
	return fmt.Errorf("Not implemented")
}
