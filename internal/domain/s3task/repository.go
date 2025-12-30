package s3task

import "context"

// 👉 No ants, no AWS here
type Uploader interface {
	Upload(ctx context.Context, task UploadTask) error
}
