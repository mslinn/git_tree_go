package internal

import (
	"sync"
	"testing"
	"time"
)

// TestThreadPoolManager_Initialization tests pool initialization
func TestThreadPoolManager_Initialization(t *testing.T) {
	pool := NewThreadPoolManager(0.75)

	if pool == nil {
		t.Fatal("Expected pool to be created")
	}

	if pool.workerCount < 1 {
		t.Errorf("Expected worker count to be at least 1, got %d", pool.workerCount)
	}

	if pool.workQueue == nil {
		t.Error("Expected work queue to be initialized")
	}
}

// TestThreadPoolManager_InvalidPercent tests that invalid percentages are rejected
func TestThreadPoolManager_InvalidPercent(t *testing.T) {
	// Test negative percent
	pool := NewThreadPoolManager(-0.5)
	if pool != nil {
		t.Error("Expected nil pool for negative percent")
	}

	// Test zero percent
	pool = NewThreadPoolManager(0)
	if pool != nil {
		t.Error("Expected nil pool for zero percent")
	}

	// Test percent > 1
	pool = NewThreadPoolManager(1.5)
	if pool != nil {
		t.Error("Expected nil pool for percent > 1")
	}
}

// TestThreadPoolManager_ValidPercents tests that valid percentages work
func TestThreadPoolManager_ValidPercents(t *testing.T) {
	validPercents := []float64{0.1, 0.5, 0.75, 1.0}

	for _, percent := range validPercents {
		pool := NewThreadPoolManager(percent)
		if pool == nil {
			t.Errorf("Expected pool to be created for percent %f", percent)
		}
	}
}

// TestThreadPoolManager_QuietVerbosity tests that initialization respects quiet verbosity
func TestThreadPoolManager_QuietVerbosity(t *testing.T) {
	// Save original verbosity
	originalVerbosity := GetVerbosity()
	defer SetVerbosity(originalVerbosity)

	// Set verbosity to QUIET
	SetVerbosity(LogQuiet)

	// Create a pool (should not log)
	pool := NewThreadPoolManager(0.75)
	if pool == nil {
		t.Fatal("Expected pool to be created")
	}

	// Start the pool (should not log initialization messages)
	pool.Start(func(task interface{}, workerID int) {
		// No-op
	})

	pool.Shutdown()
	pool.WaitForCompletion()
}

// TestThreadPoolManager_StartAndShutdown tests starting and shutting down the pool
func TestThreadPoolManager_StartAndShutdown(t *testing.T) {
	pool := NewThreadPoolManager(0.5)
	if pool == nil {
		t.Fatal("Expected pool to be created")
	}

	// Start the pool
	pool.Start(func(task interface{}, workerID int) {
		// No-op task
	})

	// Shutdown the pool
	pool.Shutdown()
	pool.WaitForCompletion()

	// Pool should have completed
	if !pool.started {
		t.Error("Expected pool to be marked as started")
	}
}

// TestThreadPoolManager_AddTask tests adding tasks to the pool
func TestThreadPoolManager_AddTask(t *testing.T) {
	pool := NewThreadPoolManager(0.5)
	if pool == nil {
		t.Fatal("Expected pool to be created")
	}

	tasksProcessed := 0
	var mu sync.Mutex

	// Start the pool
	pool.Start(func(task interface{}, workerID int) {
		mu.Lock()
		tasksProcessed++
		mu.Unlock()
	})

	// Add some tasks
	numTasks := 10
	for i := 0; i < numTasks; i++ {
		pool.AddTask(i)
	}

	// Wait for completion
	pool.WaitForCompletion()

	// All tasks should be processed
	if tasksProcessed != numTasks {
		t.Errorf("Expected %d tasks to be processed, got %d", numTasks, tasksProcessed)
	}
}

// TestThreadPoolManager_ProcessMultipleTasks tests processing multiple tasks
func TestThreadPoolManager_ProcessMultipleTasks(t *testing.T) {
	pool := NewThreadPoolManager(0.5)
	if pool == nil {
		t.Fatal("Expected pool to be created")
	}

	results := make([]int, 0)
	var mu sync.Mutex

	// Start the pool
	pool.Start(func(task interface{}, workerID int) {
		if val, ok := task.(int); ok {
			mu.Lock()
			results = append(results, val*2)
			mu.Unlock()
		}
	})

	// Add tasks
	numTasks := 20
	for i := 0; i < numTasks; i++ {
		pool.AddTask(i)
	}

	// Wait for completion
	pool.WaitForCompletion()

	// Check results
	if len(results) != numTasks {
		t.Errorf("Expected %d results, got %d", numTasks, len(results))
	}
}

// TestThreadPoolManager_ConcurrentExecution tests that tasks run concurrently
func TestThreadPoolManager_ConcurrentExecution(t *testing.T) {
	pool := NewThreadPoolManager(0.75)
	if pool == nil {
		t.Fatal("Expected pool to be created")
	}

	// Use a channel to track concurrent execution
	executing := make(chan bool, 10)
	maxConcurrent := 0
	var mu sync.Mutex

	// Start the pool
	pool.Start(func(task interface{}, workerID int) {
		executing <- true

		mu.Lock()
		current := len(executing)
		if current > maxConcurrent {
			maxConcurrent = current
		}
		mu.Unlock()

		// Simulate some work
		time.Sleep(10 * time.Millisecond)

		<-executing
	})

	// Add tasks
	numTasks := 10
	for i := 0; i < numTasks; i++ {
		pool.AddTask(i)
	}

	// Wait for completion
	pool.WaitForCompletion()

	// Should have had some concurrent execution (more than 1 task at a time)
	if maxConcurrent <= 1 {
		t.Errorf("Expected concurrent execution, but max concurrent was %d", maxConcurrent)
	}
}

// TestThreadPoolManager_WorkerIDs tests that worker IDs are unique
func TestThreadPoolManager_WorkerIDs(t *testing.T) {
	pool := NewThreadPoolManager(0.5)
	if pool == nil {
		t.Fatal("Expected pool to be created")
	}

	workerIDs := make(map[int]bool)
	var mu sync.Mutex

	// Start the pool
	pool.Start(func(task interface{}, workerID int) {
		mu.Lock()
		workerIDs[workerID] = true
		mu.Unlock()

		// Sleep to ensure multiple workers are used
		time.Sleep(5 * time.Millisecond)
	})

	// Add enough tasks to ensure multiple workers are used
	numTasks := pool.workerCount * 3
	for i := 0; i < numTasks; i++ {
		pool.AddTask(i)
	}

	// Wait for completion
	pool.WaitForCompletion()

	// Should have multiple unique worker IDs
	if len(workerIDs) < 2 {
		t.Errorf("Expected at least 2 unique worker IDs, got %d", len(workerIDs))
	}

	// All worker IDs should be in the valid range
	for id := range workerIDs {
		if id < 0 || id >= pool.workerCount {
			t.Errorf("Worker ID %d is out of range [0, %d)", id, pool.workerCount)
		}
	}
}

// TestThreadPoolManager_CannotAddTasksBeforeStart tests that adding tasks before start panics
func TestThreadPoolManager_CannotAddTasksBeforeStart(t *testing.T) {
	pool := NewThreadPoolManager(0.5)
	if pool == nil {
		t.Fatal("Expected pool to be created")
	}

	// Try to add a task before starting the pool
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when adding task before starting pool")
		}
	}()

	pool.AddTask("test")
}

// TestThreadPoolManager_DoubleStart tests that starting twice is safe
func TestThreadPoolManager_DoubleStart(t *testing.T) {
	pool := NewThreadPoolManager(0.5)
	if pool == nil {
		t.Fatal("Expected pool to be created")
	}

	// Start the pool
	pool.Start(func(task interface{}, workerID int) {
		// No-op
	})

	// Start again (should be no-op)
	pool.Start(func(task interface{}, workerID int) {
		// No-op
	})

	pool.Shutdown()
	pool.WaitForCompletion()
}

// TestThreadPoolManager_ShutdownWithoutStart tests that shutdown without start is safe
func TestThreadPoolManager_ShutdownWithoutStart(t *testing.T) {
	pool := NewThreadPoolManager(0.5)
	if pool == nil {
		t.Fatal("Expected pool to be created")
	}

	// Shutdown without starting (should be no-op)
	pool.Shutdown()
	pool.WaitForCompletion()
}

// TestThreadPoolManager_TaskOrder tests that tasks are processed (not necessarily in order, but all are processed)
func TestThreadPoolManager_TaskOrder(t *testing.T) {
	pool := NewThreadPoolManager(0.75)
	if pool == nil {
		t.Fatal("Expected pool to be created")
	}

	processed := make(map[int]bool)
	var mu sync.Mutex

	// Start the pool
	pool.Start(func(task interface{}, workerID int) {
		if val, ok := task.(int); ok {
			mu.Lock()
			processed[val] = true
			mu.Unlock()
		}
	})

	// Add tasks
	numTasks := 50
	for i := 0; i < numTasks; i++ {
		pool.AddTask(i)
	}

	// Wait for completion
	pool.WaitForCompletion()

	// Check that all tasks were processed
	for i := 0; i < numTasks; i++ {
		if !processed[i] {
			t.Errorf("Task %d was not processed", i)
		}
	}
}
