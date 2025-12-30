# Environment Variables for Testing

For integration tests with LocalStack, set these environment variables:

```bash
export AWS_ENDPOINT=http://localhost:4566
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
```

## Running Integration Tests

1. Start LocalStack:

```bash
docker-compose up -d
```

1. Run integration tests:

```bash
go test -tags=integration ./...

# Verify the file was uploaded
aws --endpoint-url=http://localhost:4566 s3 ls s3://test-bucket/
```
