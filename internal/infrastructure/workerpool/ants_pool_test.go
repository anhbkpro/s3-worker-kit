package workerpool

import (
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAntsPool_Submit(t *testing.T) {
	pool, err := NewAntsPool(2)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Release()

	var counter int64
	var wg sync.WaitGroup

	// Submit 5 tasks
	for i := 0; i < 5; i++ {
		wg.Add(1)
		err := pool.Submit(func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
			time.Sleep(10 * time.Millisecond) // Simulate work
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	wg.Wait()

	if atomic.LoadInt64(&counter) != 5 {
		t.Errorf("expected 5 tasks to complete, got %d", counter)
	}
}

func TestAntsPool_ResourceCleanup(t *testing.T) {
	// Capture log output to verify cleanup logging
	var logOutput []byte
	logWriter := &mockWriter{data: &logOutput}
	originalLogger := log.Default()
	log.SetOutput(logWriter)
	defer log.SetOutput(originalLogger.Writer())

	pool, err := NewAntsPool(1)
	if err != nil {
		t.Fatal(err)
	}

	var taskExecuted bool
	err = pool.Submit(func() {
		taskExecuted = true
		log.Println("task started")
		time.Sleep(50 * time.Millisecond)
		log.Println("task completed")
	})
	if err != nil {
		t.Fatal(err)
	}

	// Wait for task to complete
	time.Sleep(100 * time.Millisecond)

	if !taskExecuted {
		t.Error("task was not executed")
	}

	// Release pool and verify cleanup
	pool.Release()

	// Check that pool is properly cleaned up by trying to submit after release
	err = pool.Submit(func() {})
	if err == nil {
		t.Error("expected error when submitting to released pool")
	}
}

func TestAntsPool_ConcurrentTasks(t *testing.T) {
	pool, err := NewAntsPool(3)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Release()

	const numTasks = 10
	var completedTasks int64
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		err := pool.Submit(func() {
			defer wg.Done()
			atomic.AddInt64(&completedTasks, 1)
			time.Sleep(20 * time.Millisecond) // Simulate work
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	wg.Wait()
	elapsed := time.Since(start)

	if atomic.LoadInt64(&completedTasks) != numTasks {
		t.Errorf("expected %d tasks to complete, got %d", numTasks, completedTasks)
	}

	// With 3 workers and 20ms tasks, total time should be around 60-80ms (2 batches of 3)
	// Allow some tolerance for scheduling
	if elapsed > 150*time.Millisecond {
		t.Errorf("tasks took too long: %v", elapsed)
	}
}

func TestAntsPool_Release(t *testing.T) {
	pool, err := NewAntsPool(2)
	if err != nil {
		t.Fatal(err)
	}

	// Verify pool works before release
	var executed bool
	err = pool.Submit(func() {
		executed = true
	})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Millisecond) // Wait for task

	if !executed {
		t.Error("task should have executed before pool release")
	}

	// Release pool
	pool.Release()

	// Verify pool is released and doesn't accept new tasks
	err = pool.Submit(func() {})
	if err == nil {
		t.Error("pool should not accept tasks after release")
	}
}

func TestNewAntsPool_InvalidSize(t *testing.T) {
	_, err := NewAntsPool(0)
	if err == nil {
		t.Error("expected error for invalid pool size")
	}

	_, err = NewAntsPool(-1)
	if err == nil {
		t.Error("expected error for negative pool size")
	}
}

func TestAntsPool_LongRunningTasks(t *testing.T) {
	pool, err := NewAntsPool(2)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Release()

	var completed int64
	var wg sync.WaitGroup

	// Submit tasks that take different amounts of time
	tasks := []time.Duration{50 * time.Millisecond, 30 * time.Millisecond, 70 * time.Millisecond}

	for _, duration := range tasks {
		wg.Add(1)
		d := duration
		err := pool.Submit(func() {
			defer wg.Done()
			time.Sleep(d)
			atomic.AddInt64(&completed, 1)
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	wg.Wait()

	if atomic.LoadInt64(&completed) != int64(len(tasks)) {
		t.Errorf("expected %d tasks to complete, got %d", len(tasks), completed)
	}
}

// mockWriter captures log output for testing
type mockWriter struct {
	data *[]byte
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	*m.data = append(*m.data, p...)
	return len(p), nil
}
