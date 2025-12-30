# S3 Worker Kit - API Documentation

## Running the API Server

Start the HTTP server mode:

```bash
# Set LocalStack environment variables
export AWS_ENDPOINT=http://localhost:4566
export AWS_REGION=us-east-1

# Start LocalStack
docker-compose up -d

# Run the API server on port 4005
AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_ENDPOINT=http://localhost:4566 AWS_REGION=us-east-1 ./worker -mode=server -port=4005

## Graceful Shutdown

The server supports graceful shutdown with proper signal handling:

- **SIGINT (Ctrl+C)** and **SIGTERM** signals trigger graceful shutdown
- Active requests are allowed to complete (up to 30-second timeout)
- Worker pool resources are properly cleaned up
- Server shuts down cleanly without hanging
```

## Testing with REST Client

Use the provided request files with your IDE's REST Client extension:

- **`api-tests.http`** - Comprehensive test suite with all endpoints
- **`.rest`** - Simple examples in alternative format

### VS Code REST Client Setup

1. Install "REST Client" extension
2. Open `api-tests.http` file
3. Click "Send Request" above any request block

### IntelliJ HTTP Client Setup

1. Use `.rest` file format
2. Right-click on request → "Run" or "Submit Request"

## API Endpoints

### Health Check

```bash
GET /api/v1/health
```

Response:

```json
{
  "status": "healthy",
  "service": "s3-worker-kit"
}
```

### Upload Multiple Tasks (JSON)

```bash
POST /api/v1/upload/tasks
Content-Type: application/json

{
  "tasks": [
    {
      "bucket": "test-bucket",
      "key": "file1.txt",
      "data": "SGVsbG8gV29ybGQ="  // Base64 encoded content
    },
    {
      "bucket": "test-bucket",
      "key": "file2.txt",
      "data": "SG93IGFyZSB5b3U/Cg=="
    }
  ]
}
```

Response:

```json
{
  "message": "Upload completed successfully",
  "uploaded_count": 2
}
```

### Upload Single File (Multipart Form)

```bash
POST /api/v1/upload/file
Content-Type: multipart/form-data

# Form fields:
# - file: The file to upload
# - bucket: (optional) S3 bucket name, defaults to "test-bucket"
# - key: (optional) S3 key, defaults to original filename

curl -X POST \
  -F "file=@myfile.txt" \
  -F "bucket=my-bucket" \
  -F "key=custom-name.txt" \
  http://localhost:4000/api/v1/upload/file
```

Response:

```json
{
  "message": "File uploaded successfully",
  "bucket": "my-bucket",
  "key": "custom-name.txt",
  "size": 1024
}
```

## CLI Mode

For one-time uploads without starting a server:

```bash
./worker -mode=cli -bucket=test-bucket -key=myfile.txt -data="Hello World"
```

## Verification

Check uploaded files:

```bash
aws --endpoint-url=http://localhost:4566 s3 ls s3://test-bucket/
aws --endpoint-url=http://localhost:4566 s3 cp s3://test-bucket/myfile.txt -
```

## Structured Logging

All API requests include comprehensive structured logging with:

- `service_name`: Always "s3-worker-kit"
- `component`: API component (upload_middleware, ants_pool, etc.)
- Request details and performance metrics
- Error tracking with full context

Example log:

```json
{
  "timestamp": "2025-12-30T16:07:10+07:00",
  "level": "INFO",
  "msg": "starting upload",
  "service_name": "s3-worker-kit",
  "component": "upload_middleware",
  "bucket": "test-bucket",
  "key": "api-test.txt",
  "size_bytes": 23
}
```

## REST Client Testing Files

The repository includes ready-to-use HTTP request files for testing:

### `api-tests.http` - Comprehensive Test Suite

Contains examples for:

- Health checks
- Single file uploads (multipart form)
- Multiple task uploads (JSON)
- Error handling scenarios
- Large file simulations

### `.rest` - Simple Examples

Alternative format for IntelliJ HTTP Client and other tools.

### Usage with VS Code REST Client

1. Install "REST Client" extension
2. Open `api-tests.http`
3. Click "Send Request" above any request block
4. View responses and structured logs

### Usage with IntelliJ HTTP Client

1. Use `.rest` file format
2. Right-click on request → "Run" or "Submit Request"
