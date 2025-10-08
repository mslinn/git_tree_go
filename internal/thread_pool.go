package internal

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ThreadPoolManager manages a fixed-size pool of worker goroutines.
type ThreadPoolManager struct {
	workerCount int
	workQueue   chan interface{}
	workers     []*sync.WaitGroup
	mu          sync.Mutex
	started     bool
}

// NewThreadPoolManager creates a new thread pool manager.
// percentAvailableProcessors should be between 0 and 1 (e.g., 0.75 for 75%).
func NewThreadPoolManager(percentAvailableProcessors float64) *ThreadPoolManager {
	if percentAvailableProcessors <= 0 || percentAvailableProcessors > 1 {
		Log(LogQuiet, fmt.Sprintf("Error: The allowable range for ThreadPool percent_available_processors is between 0 and 1. You provided %f.", percentAvailableProcessors), ColorRed)
		return nil
	}

	workerCount := int(float64(runtime.NumCPU()) * percentAvailableProcessors)
	if workerCount < 1 {
		workerCount = 1
	}

	return &ThreadPoolManager{
		workerCount: workerCount,
		workQueue:   make(chan interface{}, 100),
		workers:     make([]*sync.WaitGroup, 0),
	}
}

// Start starts the worker pool with the given task function.
func (tp *ThreadPoolManager) Start(taskFunc func(task interface{}, workerID int)) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if tp.started {
		return
	}
	tp.started = true

	Log(LogDebug, fmt.Sprintf("Initializing %d worker threads...", tp.workerCount), ColorGreen)

	for i := 0; i < tp.workerCount; i++ {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		tp.workers = append(tp.workers, wg)

		workerID := i
		go tp.worker(workerID, taskFunc, wg)
	}
}

// worker is the worker goroutine function.
func (tp *ThreadPoolManager) worker(workerID int, taskFunc func(task interface{}, workerID int), wg *sync.WaitGroup) {
	defer wg.Done()

	Log(LogDebug, fmt.Sprintf("  [Worker %d] Started.", workerID), ColorCyan)
	startTime := time.Now()
	tasksProcessed := 0

	for task := range tp.workQueue {
		if task == nil {
			// Shutdown signal
			break
		}
		taskFunc(task, workerID)
		tasksProcessed++
	}

	if GetVerbosity() >= LogVerbose {
		elapsed := time.Since(startTime)
		msg := fmt.Sprintf("  [Worker %d] Shutting down. Processed %d tasks. Elapsed: %.2fs",
			workerID, tasksProcessed, elapsed.Seconds())
		Log(LogVerbose, msg, ColorCyan)
	}
}

// AddTask adds a task to the work queue.
func (tp *ThreadPoolManager) AddTask(task interface{}) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if !tp.started {
		panic("Cannot add tasks before the pool has been started")
	}

	tp.workQueue <- task
}

// Shutdown signals the pool to shut down gracefully.
func (tp *ThreadPoolManager) Shutdown() {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if !tp.started {
		return
	}

	// Signal all workers to shut down
	for i := 0; i < tp.workerCount; i++ {
		tp.workQueue <- nil
	}
}

// WaitForCompletion waits for all workers to complete.
func (tp *ThreadPoolManager) WaitForCompletion() {
	tp.Shutdown()

	// Wait for all workers to finish
	for _, wg := range tp.workers {
		wg.Wait()
	}

	close(tp.workQueue)
	Log(LogVerbose, "All work is complete.", ColorGreen)
}
