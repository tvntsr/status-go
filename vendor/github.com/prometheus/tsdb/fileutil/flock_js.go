// +build js

package fileutil

// ::FIXME:: it should be implemented correctly

type jsLock struct {
}

func (fl *jsLock) Release() error {
	return nil
}

func newLock(fileName string) (Releaser, error) {
	return &jsLock{}, nil
}
