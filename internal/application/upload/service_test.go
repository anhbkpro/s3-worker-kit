package upload

import (
	"context"
	"testing"

	"s3-worker-kit/internal/domain/s3task"
)

type fakeUploader struct {
	calls int
}

func (f *fakeUploader) Upload(ctx context.Context, task s3task.UploadTask) error {
	f.calls++
	return nil
}

type fakePool struct{}

func (f *fakePool) Submit(fn func()) error {
	fn()
	return nil
}
func (f *fakePool) Release() {}

func TestUploadAll(t *testing.T) {
	uploader := &fakeUploader{}
	pool := &fakePool{}

	svc := NewService(uploader, pool)

	err := svc.UploadAll(context.Background(), []s3task.UploadTask{
		{Bucket: "test-bucket", Key: "a.txt", Body: []byte("hello")},
		{Bucket: "test-bucket", Key: "b.txt", Body: []byte("world")},
		{Bucket: "test-bucket", Key: "c.txt", Body: []byte("foo")},
	})

	if err != nil {
		t.Fatal(err)
	}
	if uploader.calls != 3 {
		t.Fatalf("expected 3 uploads, got %d", uploader.calls)
	}
}
