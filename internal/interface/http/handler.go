package http

import (
	"net/http"

	"s3-worker-kit/internal/application/upload"
	"s3-worker-kit/internal/domain/s3task"
	"s3-worker-kit/internal/observability"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *upload.Service
	logger  observability.Logger
}

func NewHandler(service *upload.Service, logger observability.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger.With("component", "http_handler"),
	}
}

// UploadTasksRequest represents the JSON request for uploading multiple tasks
type UploadTasksRequest struct {
	Tasks []UploadTaskRequest `json:"tasks" binding:"required"`
}

type UploadTaskRequest struct {
	Bucket string `json:"bucket" binding:"required"`
	Key    string `json:"key" binding:"required"`
	Data   string `json:"data" binding:"required"` // Base64 encoded data
}

// UploadTasks handles JSON-based upload requests
func (h *Handler) UploadTasks(c *gin.Context) {
	var req UploadTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("received upload tasks request", "task_count", len(req.Tasks))

	tasks := make([]s3task.UploadTask, len(req.Tasks))
	for i, taskReq := range req.Tasks {
		tasks[i] = s3task.UploadTask{
			Bucket: taskReq.Bucket,
			Key:    taskReq.Key,
			Body:   []byte(taskReq.Data),
		}
	}

	if err := h.service.UploadAll(c.Request.Context(), tasks); err != nil {
		h.logger.Error("upload tasks failed", "error", err, "task_count", len(tasks))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Upload failed",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("upload tasks completed successfully", "task_count", len(tasks))
	c.JSON(http.StatusOK, gin.H{
		"message":        "Upload completed successfully",
		"uploaded_count": len(tasks),
	})
}

// UploadFile handles multipart form file uploads
func (h *Handler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Error("failed to get uploaded file", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to get uploaded file",
			"details": err.Error(),
		})
		return
	}
	defer file.Close()

	bucket := c.PostForm("bucket")
	if bucket == "" {
		bucket = "test-bucket" // Default bucket
	}

	key := c.PostForm("key")
	if key == "" {
		key = header.Filename // Use original filename as key
	}

	h.logger.Info("received file upload",
		"bucket", bucket,
		"key", key,
		"filename", header.Filename,
		"size", header.Size,
	)

	// Read file content
	content := make([]byte, header.Size)
	_, err = file.Read(content)
	if err != nil {
		h.logger.Error("failed to read file content", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to read file content",
			"details": err.Error(),
		})
		return
	}

	task := s3task.UploadTask{
		Bucket: bucket,
		Key:    key,
		Body:   content,
	}

	if err := h.service.UploadAll(c.Request.Context(), []s3task.UploadTask{task}); err != nil {
		h.logger.Error("file upload failed", "error", err, "bucket", bucket, "key", key)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "File upload failed",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("file upload completed successfully", "bucket", bucket, "key", key)
	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"bucket":  bucket,
		"key":     key,
		"size":    header.Size,
	})
}

// HealthCheck provides a simple health check endpoint
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "s3-worker-kit",
	})
}

// SetupRoutes configures the Gin router with all routes
func SetupRoutes(handler *Handler, logger observability.Logger) *gin.Engine {
	if logger != nil {
		gin.DefaultWriter = &logWriter{logger: logger.With("component", "gin")}
	}

	router := gin.Default()

	// Add service name to all responses
	router.Use(func(c *gin.Context) {
		c.Header("X-Service-Name", observability.ServiceName)
		c.Next()
	})

	// Routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", handler.HealthCheck)
		v1.POST("/upload/tasks", handler.UploadTasks)
		v1.POST("/upload/file", handler.UploadFile)
	}

	return router
}

// logWriter implements io.Writer to redirect Gin's logs to structured logging
type logWriter struct {
	logger observability.Logger
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.logger.Info("gin", "message", string(p))
	return len(p), nil
}
