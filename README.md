# 1️⃣ High-level Architecture

## Goals

- Deterministic concurrency
- Safe S3 usage (retry, throttling, cancellation)
- Testable without AWS
- Replaceable worker pool (ants ↔ errgroup)
- Observable & operable

## Layered Design (Clean Architecture)

```text
┌──────────────────────────┐
│        Interface         │  ← CLI / HTTP / Cron
└────────────┬─────────────┘
             │
┌────────────▼─────────────┐
│        Application       │  ← Use cases
└────────────┬─────────────┘
             │
┌────────────▼─────────────┐
│          Domain          │  ← Entities, ports
└────────────┬─────────────┘
             │
┌────────────▼─────────────┐
│      Infrastructure      │  ← AWS S3, ants
└──────────────────────────┘
```

## 2️⃣ Project Structure

```text
internal/
├── domain/
│   ├── s3task/
│   │   ├── task.go
│   │   ├── repository.go
│   │   └── errors.go
│
├── application/
│   ├── upload/
│   │   ├── usecase.go
│   │   ├── service.go
│   │   └── metrics.go
│
├── infrastructure/
│   ├── s3/
│   │   ├── client.go
│   │   ├── uploader.go
│   │   └── uploader_test.go
│   │
│   ├── workerpool/
│   │   ├── pool.go
│   │   ├── ants_pool.go
│   │   └── ants_pool_test.go
│
├── config/
│   └── config.go
│
├── observability/
│   ├── logger.go
│   └── metrics.go
│
cmd/
└── worker/
    └── main.go
```

## 3️⃣ Domain Layer (Pure, No AWS)

- Task Entity

```go
// domain/s3task/task.go
package s3task

type UploadTask struct {
  Bucket string
  Key    string
  Body   []byte
}
```

- Port (Repository Interface)

```go
// domain/s3task/repository.go
package s3task

import "context"

type Uploader interface {
  Upload(ctx context.Context, task UploadTask) error
}
```

👉 No ants, no AWS here

## 4️⃣ Worker Pool Abstraction (Critical)

Port

```go
// infrastructure/workerpool/pool.go
package workerpool

type Pool interface {
  Submit(func()) error
  Release()
}
```

- ants Implementation

```go
// infrastructure/workerpool/ants_pool.go
package workerpool

import "github.com/panjf2000/ants/v2"

type AntsPool struct {
  pool *ants.Pool
}

func NewAntsPool(size int) (*AntsPool, error) {
  p, err := ants.NewPool(size, ants.WithPreAlloc(true))
  if err != nil {
    return nil, err
  }
  return &AntsPool{pool: p}, nil
}

func (a *AntsPool) Submit(fn func()) error {
  return a.pool.Submit(fn)
}

func (a *AntsPool) Release() {
  a.pool.Release()
}
```

## 5️⃣ Infrastructure: S3 Implementation

- Client Setup (Retry + Timeouts)

```go
// infrastructure/s3/client.go
package s3infra

import (
  "context"
  "time"

  "github.com/aws/aws-sdk-go-v2/aws"
  "github.com/aws/aws-sdk-go-v2/config"
  "github.com/aws/aws-sdk-go-v2/retry"
  "github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewS3Client(ctx context.Context) (*s3.Client, error) {
  cfg, err := config.LoadDefaultConfig(ctx,
    config.WithRetryer(func() aws.Retryer {
      return retry.NewStandard(func(o *retry.StandardOptions) {
        o.MaxAttempts = 5
      })
    }),
  )
  if err != nil {
    return nil, err
  }
  return s3.NewFromConfig(cfg), nil
}
```

- Uploader Implementation

```go
// infrastructure/s3/uploader.go
package s3infra

import (
  "bytes"
  "context"

  "github.com/aws/aws-sdk-go-v2/service/s3"
  "s3-worker-kit/internal/domain/s3task"
)

type Uploader struct {
  client *s3.Client
}

func NewUploader(client *s3.Client) *Uploader {
  return &Uploader{client: client}
}

func (u *Uploader) Upload(ctx context.Context, task s3task.UploadTask) error {
  _, err := u.client.PutObject(ctx, &s3.PutObjectInput{
    Bucket: &task.Bucket,
    Key:    &task.Key,
    Body:   bytes.NewReader(task.Body),
  })
  return err
}
```

## 6️⃣ Application Layer (Use Case)

Upload Service

```go
// application/upload/service.go
package upload

import (
  "context"
  "sync"

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
```

Use Case Execution

```go
// application/upload/usecase.go
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
```

## 7️⃣ Testing Strategy (Very Important)

A. Unit Tests (No AWS, No ants)
Fake Uploader

```go
type FakeUploader struct {
  Calls int
  Err   error
}

func (f *FakeUploader) Upload(ctx context.Context, _ s3task.UploadTask) error {
  f.Calls++
  return f.Err
}
```

Fake Pool (Runs synchronously)

```go
type FakePool struct{}

func (f *FakePool) Submit(fn func()) error {
  fn()
  return nil
}
func (f *FakePool) Release() {}
```

Test Use Case

```go
func TestUploadAll_Success(t *testing.T) {
  uploader := &FakeUploader{}
  pool := &FakePool{}

  svc := upload.NewService(uploader, pool)

  tasks := []s3task.UploadTask{{}, {}, {}}

  err := svc.UploadAll(context.Background(), tasks)
  assert.NoError(t, err)
  assert.Equal(t, 3, uploader.Calls)
}
```

B. Integration Tests (Localstack / MinIO)

- Run S3 locally
- Real AWS SDK
- Small concurrency

C. Concurrency Tests

```go
go test -race ./...
```

Mandatory for this system.

## 8️⃣ Observability (Production-grade)

Metrics to expose

- s3_upload_success_total
- s3_upload_error_total
- s3_upload_latency_ms
- worker_pool_inflight

Example Hook

```go
start := time.Now()
err := s.uploader.Upload(ctx, t)
metrics.UploadLatency.Observe(time.Since(start).Seconds())
```

## 9️⃣ Operational Concerns

Backpressure

- ants pool limits concurrency
- submit error → fast fail

Graceful Shutdown

```go
ctx, cancel := context.WithCancel(context.Background())
signal.Notify(sig, os.Interrupt)
go func() {
  <-sig
  cancel()
}()
```

Retry

- SDK retry only
- Do NOT retry at worker layer

## Development

### Prerequisites

- Go 1.21+
- Docker & Docker Compose (for LocalStack)
- Git

### Setup

1. Clone the repository
2. Install dependencies: `go mod download`
3. Start LocalStack: `docker-compose up -d`
4. Create test bucket: `aws --endpoint-url=http://localhost:4566 s3 mb s3://test-bucket`

### Running

- **Build**: `go build -o worker cmd/worker/main.go`
- **Run**: `AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_ENDPOINT=http://localhost:4566 AWS_REGION=us-east-1 ./worker -mode=server -port=4006`

- **CLI Mode**: `go run cmd/worker/main.go -mode=cli`
- **Server Mode**: `go run cmd/worker/main.go -mode=server -port=8080`
- **Tests**: `go test ./...`

### .gitignore

The project includes a comprehensive `.gitignore` file that ignores:

- **Go artifacts**: Binaries, test files, coverage reports
- **Dependencies**: Vendor directory, module cache
- **IDE files**: VS Code, IntelliJ, Vim swap files
- **OS files**: macOS `.DS_Store`, Windows `Thumbs.db`
- **Environment**: `.env` files (use `.env.example` as template)
- **Logs**: Log files and temporary directories
- **Build artifacts**: `dist/`, `build/`, `bin/` directories
- **Test files**: Generated test files like `*_test.txt`

This ensures a clean repository with only source code and documentation.

5️⃣ Infrastructure: S3 Implementation

6️⃣ Application Layer (Use Case)

- Upload Service

```go
// application/upload/service.go
package upload

import (
  "context"
  "sync"

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
```
