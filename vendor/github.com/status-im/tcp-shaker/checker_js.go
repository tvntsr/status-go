// +build js

// ::FIXME:: all method should be implemented
package tcp

import (
	"fmt"
	"context"
	"time"
)

// Checker contains an epoll instance for TCP handshake checking
type Checker struct {
	pipePool
	resultPipes
}

// NewChecker creates a Checker with linger set to zero.
func NewChecker() *Checker {
	return NewCheckerZeroLinger(true)
}

// NewCheckerZeroLinger creates a Checker with zeroLinger set to given value.
func NewCheckerZeroLinger(zeroLinger bool) *Checker {
	return &Checker{
		pipePool:    newPipePoolSyncPool(),
		resultPipes: newResultPipesSyncMap(),
	}
}

// CheckingLoop must be called before anything else.
// NOTE: this function blocks until ctx got canceled.
func (c *Checker) CheckingLoop(ctx context.Context) error {
	return fmt.Errorf("CheckingLoop is not implemented")
}


const pollerTimeout = time.Second


func (c *Checker) PollerFDAtomic() int {
	return int(0)
}

// CheckAddr performs a TCP check with given TCP address and timeout
// A successful check will result in nil error
// ErrTimeout is returned if timeout
// zeroLinger is an optional parameter indicating if linger should be set to zero
// for this particular connection
// Note: timeout includes domain resolving
func (c *Checker) CheckAddr(addr string, timeout time.Duration) (err error) {
	return fmt.Errorf("Not implemented")
}

// CheckAddrZeroLinger is like CheckAddr with an extra parameter indicating whether to enable zero linger.
func (c *Checker) CheckAddrZeroLinger(addr string, timeout time.Duration, zeroLinger bool) error {
	return fmt.Errorf("Not implemented")
}

// WaitReady returns a chan which is closed when the Checker is ready for use.
func (c *Checker) WaitReady() <-chan struct{} {
	panic("Not implemented")
}

// IsReady returns a bool indicates whether the Checker is ready for use
func (c *Checker) IsReady() bool {
	return c.PollerFDAtomic() > 0
}

// PollerFd returns the inner fd of poller instance.
// NOTE: Use this only when you really know what you are doing.
func (c *Checker) PollerFd() int {
	return c.PollerFDAtomic()
}
