package upload

import (
	"context"
	"sync"
	"time"

	"s3-worker-kit/internal/domain/s3task"
	"s3-worker-kit/internal/observability"
)

func (s *Service) UploadAll(
	ctx context.Context,
	tasks []s3task.UploadTask,
) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(tasks))

	for _, task := range tasks {
		t := task
		wg.Add(1)

		err := s.pool.Submit(func() {
			defer wg.Done()

			// Track active uploads
			observability.ActiveUploads.Inc()
			defer observability.ActiveUploads.Dec()

			// Measure upload latency
			start := time.Now()
			err := s.uploader.Upload(ctx, t)
			duration := time.Since(start).Seconds()

			// Record metrics
			status := "success"
			if err != nil {
				status = "error"
				errCh <- err
			}

			observability.UploadLatency.WithLabelValues(t.Bucket, status).Observe(duration)
			observability.UploadTotal.WithLabelValues(t.Bucket, status).Inc()
		})
		if err != nil {
			wg.Done()
			return err
		}
	}

	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		return <-errCh // or aggregate
	}
	return nil
}
