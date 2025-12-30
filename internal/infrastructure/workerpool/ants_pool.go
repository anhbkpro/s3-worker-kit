package workerpool

import (
	"s3-worker-kit/internal/observability"

	"github.com/panjf2000/ants/v2"
)

type AntsPool struct {
	pool   *ants.Pool
	logger observability.Logger
}

func NewAntsPool(size int) (*AntsPool, error) {
	logger := observability.NewDefaultLogger().With("component", "ants_pool")

	logger.Info("creating ants pool", "size", size)

	p, err := ants.NewPool(size, ants.WithPreAlloc(true))
	if err != nil {
		logger.Error("failed to create ants pool", "error", err, "size", size)
		return nil, err
	}

	logger.Info("ants pool created successfully", "size", size)

	return &AntsPool{
		pool:   p,
		logger: logger,
	}, nil
}

func (a *AntsPool) Submit(fn func()) error {
	a.logger.Debug("submitting task to ants pool")

	err := a.pool.Submit(func() {
		a.logger.Debug("task started execution")
		defer a.logger.Debug("task completed")
		fn()
	})

	if err != nil {
		a.logger.Error("failed to submit task", "error", err)
		return err
	}

	a.logger.Debug("task submitted successfully")
	return nil
}

func (a *AntsPool) Release() {
	a.logger.Info("releasing ants pool resources")
	a.pool.Release()
	a.logger.Info("ants pool resources released")
}
