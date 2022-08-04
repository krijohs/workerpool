package workerpool

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
)

type Pool[T any] struct {
	ctx       context.Context
	cancel    context.CancelFunc
	workerWg  *sync.WaitGroup
	jobsWg    *sync.WaitGroup
	mu        sync.RWMutex
	options   *Options
	isRunning bool
	jobCh     chan JobFn[T]
	resultCh  chan JobResult[T]
}

type JobResult[T any] struct {
	Result T
	Err    error
}

type JobFn[T any] func() (T, error)

// New initalizes a new pool with optional options.
func New[T any](opts ...Option) *Pool[T] {
	p := &Pool[T]{
		workerWg: &sync.WaitGroup{},
		jobsWg:   &sync.WaitGroup{},
	}

	options := &Options{
		workers:        runtime.NumCPU(),
		jobBufferSize:  10,
		disableResults: false,
	}

	for _, opt := range opts {
		opt(options)
	}

	p.options = options
	p.jobCh = make(chan JobFn[T], options.jobBufferSize)
	p.resultCh = make(chan JobResult[T], options.jobBufferSize+2)

	return p
}

// Start will fire up the workers so they are ready to process incoming jobs.
func (p *Pool[T]) Start(ctx context.Context) {
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.setRunning(true)

	p.workerWg.Add(p.options.workers)
	for i := 0; i < p.options.workers; i++ {
		go func() {
			defer p.workerWg.Done()
			p.worker()
		}()
	}
}

func (p *Pool[T]) setRunning(isRunning bool) {
	p.mu.Lock()
	p.isRunning = isRunning
	p.mu.Unlock()
}

func (p *Pool[T]) worker() {
	for {
		select {
		case job, open := <-p.jobCh:
			if !open {
				return
			}

			res, err := job()
			if !p.options.disableResults {
				p.resultCh <- JobResult[T]{Result: res, Err: err}
			}
			p.jobsWg.Done()

		case <-p.ctx.Done():
			return
		}
	}
}

// Add a new job to the pool.
func (p *Pool[T]) Add(ctx context.Context, fn JobFn[T]) error {
	if err := p.checkRunning(); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.ctx.Done():
		return fmt.Errorf("pool context cancelled: %w", p.ctx.Err())
	default:
		p.jobsWg.Add(1)
		p.jobCh <- fn
	}

	return nil
}

func (p *Pool[T]) checkRunning() error {
	p.mu.RLock()
	if !p.isRunning {
		return errors.New("pool is not running")
	}
	p.mu.RUnlock()

	return nil
}

// Results are returned in a channel. Will return an error if `DisableResults` option was used.
func (p *Pool[T]) Results() (chan JobResult[T], error) {
	if err := p.checkRunning(); err != nil {
		return nil, err
	}

	if p.options.disableResults {
		return nil, errors.New("results disabled for pool")
	}

	return p.resultCh, nil
}

// Wait will wait for the jobs in the pool to be processed and then return.
func (p *Pool[T]) Wait(ctx context.Context) error {
	if err := p.checkRunning(); err != nil {
		return err
	}

	doneCh := make(chan struct{})
	go func() {
		p.jobsWg.Wait()
		p.cancel()
		close(p.jobCh)
		close(p.resultCh)
		p.workerWg.Wait()
		doneCh <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-doneCh:
		p.setRunning(false)
		return nil
	}
}
