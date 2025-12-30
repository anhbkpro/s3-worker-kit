package s3infra

import (
	"bytes"
	"context"

	"s3-worker-kit/internal/domain/s3task"

	"github.com/aws/aws-sdk-go-v2/service/s3"
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
