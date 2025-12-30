package observability

import (
	"context"
	"time"

	"s3-worker-kit/internal/domain/s3task"
)

// LoggingMiddleware provides logging functionality for services
type LoggingMiddleware struct {
	logger Logger
	next   s3task.Uploader
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger Logger, uploader s3task.Uploader) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger.With("component", "upload_middleware"),
		next:   uploader,
	}
}

// Upload logs the upload operation
func (m *LoggingMiddleware) Upload(ctx context.Context, task s3task.UploadTask) error {
	start := time.Now()

	m.logger.Info("starting upload",
		"bucket", task.Bucket,
		"key", task.Key,
		"size_bytes", len(task.Body),
	)

	err := m.next.Upload(ctx, task)

	duration := time.Since(start)

	if err != nil {
		m.logger.Error("upload failed",
			"bucket", task.Bucket,
			"key", task.Key,
			"duration_ms", duration.Milliseconds(),
			"error", err.Error(),
		)
		return err
	}

	m.logger.Info("upload completed successfully",
		"bucket", task.Bucket,
		"key", task.Key,
		"duration_ms", duration.Milliseconds(),
	)

	return nil
}

// WorkerPool defines the worker pool interface for middleware
type WorkerPool interface {
	Submit(func()) error
	Release()
}

// WorkerPoolLoggingMiddleware provides logging for worker pool operations
type WorkerPoolLoggingMiddleware struct {
	logger Logger
	next   WorkerPool
}

// Ensure WorkerPoolLoggingMiddleware implements WorkerPool interface
var _ WorkerPool = (*WorkerPoolLoggingMiddleware)(nil)

// NewWorkerPoolLoggingMiddleware creates logging middleware for worker pools
func NewWorkerPoolLoggingMiddleware(logger Logger, pool WorkerPool) *WorkerPoolLoggingMiddleware {
	return &WorkerPoolLoggingMiddleware{
		logger: logger.With("component", "worker_pool_middleware"),
		next:   pool,
	}
}

// Submit logs task submission
func (m *WorkerPoolLoggingMiddleware) Submit(fn func()) error {
	m.logger.Debug("submitting task to worker pool")

	start := time.Now()
	err := m.next.Submit(fn)

	if err != nil {
		m.logger.Error("failed to submit task to worker pool",
			"error", err.Error(),
		)
	} else {
		m.logger.Debug("task submitted to worker pool",
			"submit_duration_ms", time.Since(start).Milliseconds(),
		)
	}

	return err
}

// Release logs pool resource cleanup
func (m *WorkerPoolLoggingMiddleware) Release() {
	m.logger.Info("releasing worker pool resources")
	start := time.Now()

	m.next.Release()

	m.logger.Info("worker pool resources released",
		"release_duration_ms", time.Since(start).Milliseconds(),
	)
}
