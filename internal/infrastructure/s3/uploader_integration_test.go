package s3infra

import (
	"context"
	"os"
	"testing"

	"s3-worker-kit/internal/domain/s3task"
)

func TestUploader_Localstack(t *testing.T) {
	if os.Getenv("AWS_ENDPOINT") == "" {
		t.Skip("AWS_ENDPOINT not set")
	}

	ctx := context.Background()

	client, err := NewClient(
		ctx,
		os.Getenv("AWS_REGION"),
		os.Getenv("AWS_ENDPOINT"),
	)
	if err != nil {
		t.Fatal(err)
	}

	uploader := NewUploader(client)

	err = uploader.Upload(ctx, s3task.UploadTask{
		Bucket: "test-bucket",
		Key:    "integration.txt",
		Body:   []byte("integration test"),
	})
	if err != nil {
		t.Fatal(err)
	}
}
