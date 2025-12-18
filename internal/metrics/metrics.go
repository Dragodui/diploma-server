package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	HttpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	HttpResponseSizeBytes = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
		},
		[]string{"method", "endpoint"},
	)

	// Database Metrics
	DbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	DbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	DbConnectionsOpen = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_open",
			Help: "Number of open database connections",
		},
	)

	DbConnectionsInUse = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_in_use",
			Help: "Number of database connections currently in use",
		},
	)

	DbErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation", "table"},
	)

	// Redis Metrics
	RedisOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_operations_total",
			Help: "Total number of Redis operations",
		},
		[]string{"operation", "status"},
	)

	RedisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_operation_duration_seconds",
			Help:    "Redis operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	RedisCacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_cache_hits_total",
			Help: "Total number of Redis cache hits",
		},
	)

	RedisCacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_cache_misses_total",
			Help: "Total number of Redis cache misses",
		},
	)

	// Authentication Metrics
	AuthAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"method", "status"},
	)

	AuthTokensGenerated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_tokens_generated_total",
			Help: "Total number of JWT tokens generated",
		},
	)

	AuthTokensValidated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_tokens_validated_total",
			Help: "Total number of JWT token validation attempts",
		},
		[]string{"status"},
	)

	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of currently active users",
		},
	)

	// S3 Metrics
	S3UploadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "s3_uploads_total",
			Help: "Total number of S3 uploads",
		},
		[]string{"status"},
	)

	S3UploadDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "s3_upload_duration_seconds",
			Help:    "S3 upload duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	S3UploadSizeBytes = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "s3_upload_size_bytes",
			Help:    "S3 upload file size in bytes",
			Buckets: []float64{1024, 10240, 102400, 1048576, 10485760, 104857600},
		},
	)

	// Business Metrics - Homes
	HomesTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "homes_total",
			Help: "Total number of homes in the system",
		},
	)

	HomeMembersTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "home_members_total",
			Help: "Total number of members per home",
		},
		[]string{"home_id"},
	)

	HomeOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "home_operations_total",
			Help: "Total number of home operations",
		},
		[]string{"operation"},
	)

	// Business Metrics - Tasks
	TasksTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tasks_total",
			Help: "Total number of tasks by status",
		},
		[]string{"status"},
	)

	TaskOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_operations_total",
			Help: "Total number of task operations",
		},
		[]string{"operation"},
	)

	// Business Metrics - Bills
	BillsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "bills_total",
			Help: "Total number of bills",
		},
	)

	BillOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bill_operations_total",
			Help: "Total number of bill operations",
		},
		[]string{"operation"},
	)

	// Business Metrics - Shopping
	ShoppingItemsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "shopping_items_total",
			Help: "Total number of shopping items",
		},
	)

	ShoppingOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shopping_operations_total",
			Help: "Total number of shopping operations",
		},
		[]string{"operation"},
	)

	// Business Metrics - Polls
	PollsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "polls_total",
			Help: "Total number of polls",
		},
	)

	PollVotesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "poll_votes_total",
			Help: "Total number of poll votes",
		},
	)

	// Business Metrics - Notifications
	NotificationsSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_sent_total",
			Help: "Total number of notifications sent",
		},
		[]string{"type"},
	)

	// Email Metrics
	EmailsSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "emails_sent_total",
			Help: "Total number of emails sent",
		},
		[]string{"type", "status"},
	)

	EmailSendDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "email_send_duration_seconds",
			Help:    "Email send duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// OCR Metrics
	OcrRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ocr_requests_total",
			Help: "Total number of OCR requests",
		},
		[]string{"status"},
	)

	OcrProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ocr_processing_duration_seconds",
			Help:    "OCR processing duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
	)

	// Application Info
	AppInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_info",
			Help: "Application information",
		},
		[]string{"version", "go_version"},
	)

	// Server Uptime
	ServerStartTime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "server_start_time_seconds",
			Help: "Server start time in Unix seconds",
		},
	)
)
