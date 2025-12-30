package upload

import (
	"s3-worker-kit/internal/domain/s3task"
	"s3-worker-kit/internal/infrastructure/workerpool"
)

type Service struct {
	uploader s3task.Uploader
	pool     workerpool.Pool
}

func NewService(
	uploader s3task.Uploader,
	pool workerpool.Pool,
) *Service {
	return &Service{uploader, pool}
}
