package upload

import (
	"context"
	"s3-worker-kit/internal/domain/s3task"
	"sync"
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
			if err := s.uploader.Upload(ctx, t); err != nil {
				errCh <- err
			}
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
