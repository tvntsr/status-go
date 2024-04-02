// +build wasm

package fileutil


import (
)

// Releaser provides the Release method to release a file lock.
type Releaser interface {
        Release() error
}

// Flock locks the file with the provided name. If the file does not exist, it is
// created. The returned Releaser is used to release the lock. existed is true
// if the file to lock already existed. A non-nil error is returned if the
// locking has failed. Neither this function nor the returned Releaser is
// goroutine-safe.
func Flock(fileName string) (r Releaser, existed bool, err error) {
        r, err = newLock(fileName)
        return r, false, err
}

type jsLock struct {
}

func (fl *jsLock) Release() error {
	return nil
}

func newLock(fileName string) (Releaser, error) {
	return &jsLock{}, nil
}
