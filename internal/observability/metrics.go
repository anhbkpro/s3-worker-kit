package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// UploadLatency measures the time taken for S3 upload operations
	UploadLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "s3_worker_upload_duration_seconds",
			Help:    "Time taken to upload files to S3",
			Buckets: prometheus.DefBuckets, // 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"bucket", "status"}, // labels for bucket name and success/failure status
	)

	// UploadTotal counts the total number of upload operations
	UploadTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "s3_worker_upload_total",
			Help: "Total number of upload operations",
		},
		[]string{"bucket", "status"}, // labels for bucket name and success/failure status
	)

	// ActiveUploads tracks the number of currently active uploads
	ActiveUploads = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "s3_worker_active_uploads",
			Help: "Number of currently active upload operations",
		},
	)
)

// init function to ensure metrics are registered
func init() {
	// Force registration by accessing the metrics
	_ = UploadLatency
	_ = UploadTotal
	_ = ActiveUploads
}
