package s3task

type UploadTask struct {
	Bucket string
	Key    string
	Body   []byte
}
