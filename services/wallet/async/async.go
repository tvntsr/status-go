package async

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

type Command func(context.Context) error

// FiniteCommand terminates when error is nil.
type FiniteCommand struct {
	Interval time.Duration
	Runable  func(context.Context) error
}

func (c FiniteCommand) Run(ctx context.Context) error {
	err := c.Runable(ctx)
	if err == nil {
		return nil
	}
	ticker := time.NewTicker(c.Interval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			err := c.Runable(ctx)
			if err == nil {
				return nil
			}
		}
	}
}

// FiniteCommandWithBackoff terminates when error is nil.
// BackoffInterval is used when set, otherwise Interval is used.
type FiniteCommandWithBackoff struct {
	Interval        time.Duration
	backoffInterval time.Duration
	Runable         func(context.Context) error
}

func (c *FiniteCommandWithBackoff) Run(ctx context.Context) error {
	err := c.Runable(ctx)
	if err == nil {
		return nil
	}

	var interval time.Duration
	if c.backoffInterval != 0 {
		interval = c.backoffInterval
		c.backoffInterval = 0
	} else {
		interval = c.Interval
	}

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			err := c.Runable(ctx)
			if err == nil {
				return nil
			}
		}
	}
}

func (c *FiniteCommandWithBackoff) Backoff(timeout time.Duration) {
	c.backoffInterval = timeout
}

// InfiniteCommand runs until context is closed.
type InfiniteCommand struct {
	Interval time.Duration
	Runable  func(context.Context) error
}

func (c InfiniteCommand) Run(ctx context.Context) error {
	_ = c.Runable(ctx)
	ticker := time.NewTicker(c.Interval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			_ = c.Runable(ctx)
		}
	}
}

func NewGroup(parent context.Context) *Group {
	ctx, cancel := context.WithCancel(parent)
	return &Group{
		ctx:    ctx,
		cancel: cancel,
	}
}

type Group struct {
	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup
}

func (g *Group) Add(cmd Command) {
	g.wg.Add(1)
	go func() {
		_ = cmd(g.ctx)
		g.wg.Done()
	}()
}

func (g *Group) Stop() {
	g.cancel()
}

func (g *Group) Wait() {
	g.wg.Wait()
}

func (g *Group) WaitAsync() <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		g.Wait()
		close(ch)
	}()
	return ch
}

func NewAtomicGroup(parent context.Context) *AtomicGroup {
	ctx, cancel := context.WithCancel(parent)
	return &AtomicGroup{ctx: ctx, cancel: cancel}
}

// AtomicGroup terminates as soon as first goroutine terminates..
type AtomicGroup struct {
	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup

	mu    sync.Mutex
	error error
}

// Go spawns function in a goroutine and stores results or errors.
func (d *AtomicGroup) Add(cmd Command) {
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		err := cmd(d.ctx)
		d.mu.Lock()
		defer d.mu.Unlock()
		if err != nil {
			// do not overwrite original error by context errors
			if d.error != nil {
				return
			}
			d.error = err
			d.cancel()
			return
		}
	}()
}

// Wait for all downloaders to finish.
func (d *AtomicGroup) Wait() {
	d.wg.Wait()
	if d.Error() == nil {
		d.mu.Lock()
		defer d.mu.Unlock()
		d.cancel()
	}
}

func (d *AtomicGroup) WaitAsync() <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		d.Wait()
		close(ch)
	}()
	return ch
}

// Error stores an error that was reported by any of the downloader. Should be called after Wait.
func (d *AtomicGroup) Error() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.error
}

func (d *AtomicGroup) Stop() {
	d.cancel()
}

func NewQueuedAtomicGroup(parent context.Context, limit int64) *QueuedAtomicGroup {
	return &QueuedAtomicGroup{NewAtomicGroup(parent), limit, 0, []Command{}}
}

type QueuedAtomicGroup struct {
	*AtomicGroup
	limit       int64
	count       int64
	pendingCmds []Command
}

func (d *QueuedAtomicGroup) Add(cmd Command) {
	counter := atomic.LoadInt64(&d.count)
	if d.limit > 0 && counter >= d.limit {
		d.pendingCmds = append(d.pendingCmds, cmd)
		log.Info("queueing command", "pending", len(d.pendingCmds))
		return
	}

	d.run(cmd)
}

func (d *QueuedAtomicGroup) run(cmd Command) {
	d.wg.Add(1)
	atomic.AddInt64(&d.count, int64(1))
	go func() {
		defer d.Done()
		err := cmd(d.ctx)
		d.mu.Lock()
		defer d.mu.Unlock()
		if err != nil {
			// do not overwrite original error by context errors
			if d.error != nil {
				return
			}
			d.error = err
			d.cancel()
			return
		}
	}()
}

func (d *QueuedAtomicGroup) Done() {
	atomic.AddInt64(&d.count, -1)
	d.wg.Done()

	counter := atomic.LoadInt64(&d.count)

	log.Info("queued command done", "counter", counter, "pending len", len(d.pendingCmds))

	if counter < d.limit && len(d.pendingCmds) > 0 {
		cmd := d.pendingCmds[0]
		d.pendingCmds = d.pendingCmds[1:]
		log.Info("running queued command", "pending len", len(d.pendingCmds))
		d.run(cmd)
	}
}
