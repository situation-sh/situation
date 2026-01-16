package utils

import (
	"errors"
	"sync"
)

// WorkerPool is a generic worker pool that processes items of type T
// concurrently using a configurable number of workers.
type WorkerPool[T any] struct {
	poolSize uint
	worker   func(items <-chan T, errChan chan<- error, wg *sync.WaitGroup)
}

// NewWorkerPool creates a new worker pool with the given pool size and
// processing function. The function is called for each item in the pool.
func NewWorkerPool[T any](poolSize uint, fun func(T) error) *WorkerPool[T] {
	worker := func(itemChan <-chan T, errChan chan<- error, wg *sync.WaitGroup) {
		defer wg.Done()
		for item := range itemChan {
			if err := fun(item); err != nil {
				errChan <- err
			}
		}
	}
	return &WorkerPool[T]{
		poolSize: poolSize,
		worker:   worker,
	}
}

// Run processes all items in the given slice using the worker pool.
// It returns an aggregated error of all errors encountered during processing.
func (wp *WorkerPool[T]) Run(items []T) error {
	var wg sync.WaitGroup

	itemChan := make(chan T, 2*wp.poolSize)
	errChan := make(chan error)
	errResult := make(chan error)

	// Error consumer
	go func() {
		var errs error
		for err := range errChan {
			errs = errors.Join(errs, err)
		}
		errResult <- errs
	}()

	// start workers
	for i := uint(0); i < wp.poolSize; i++ {
		wg.Add(1)
		go wp.worker(itemChan, errChan, &wg)
	}
	// send jobs
	for _, item := range items {
		itemChan <- item
	}
	close(itemChan)
	// wait for workers to finish
	wg.Wait()
	// close errors
	close(errChan)
	return <-errResult
}
