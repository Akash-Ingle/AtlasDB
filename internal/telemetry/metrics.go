package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	EventsIngestedTotal *prometheus.CounterVec
	EventsIngestedBytes prometheus.Counter
	IngestBatchSize     prometheus.Histogram

	EventsProcessedTotal *prometheus.CounterVec
	ProcessingDuration   prometheus.Histogram
	ProcessingErrors     *prometheus.CounterVec

	QueueDepth             *prometheus.GaugeVec
	QueueOldestMessageAge  *prometheus.GaugeVec
	DLQMessagesTotal       prometheus.Counter

	DBQueryDuration    *prometheus.HistogramVec
	DBConnectionsActive prometheus.Gauge
	DBConnectionsIdle   prometheus.Gauge

	CacheHitsTotal   *prometheus.CounterVec
	CacheMissesTotal *prometheus.CounterVec

	SearchQueriesTotal  *prometheus.CounterVec
	SearchDuration      *prometheus.HistogramVec
}

func NewMetrics(namespace string) *Metrics {
	return &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		}, []string{"method", "route", "status"}),

		HTTPRequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}, []string{"method", "route"}),

		HTTPRequestsInFlight: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "http_requests_in_flight",
			Help:      "Number of HTTP requests currently being processed",
		}),

		EventsIngestedTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "events_ingested_total",
			Help:      "Total number of events ingested",
		}, []string{"source"}),

		EventsIngestedBytes: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "events_ingested_bytes_total",
			Help:      "Total bytes of ingested event data",
		}),

		IngestBatchSize: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "ingest_batch_size",
			Help:      "Number of events per ingest batch",
			Buckets:   []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
		}),

		EventsProcessedTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "events_processed_total",
			Help:      "Total number of events processed",
		}, []string{"status"}),

		ProcessingDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "processing_duration_seconds",
			Help:      "Event processing duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		}),

		ProcessingErrors: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "processing_errors_total",
			Help:      "Total number of processing errors",
		}, []string{"stage"}),

		QueueDepth: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "queue_depth",
			Help:      "Number of messages in queue",
		}, []string{"stream"}),

		QueueOldestMessageAge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "queue_oldest_message_age_seconds",
			Help:      "Age of the oldest message in the queue",
		}, []string{"stream"}),

		DLQMessagesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "dlq_messages_total",
			Help:      "Total messages sent to dead-letter queue",
		}),

		DBQueryDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "db_query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5},
		}, []string{"query"}),

		DBConnectionsActive: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_connections_active",
			Help:      "Number of active database connections",
		}),

		DBConnectionsIdle: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_connections_idle",
			Help:      "Number of idle database connections",
		}),

		CacheHitsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "cache_hits_total",
			Help:      "Total cache hits",
		}, []string{"cache"}),

		CacheMissesTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "cache_misses_total",
			Help:      "Total cache misses",
		}, []string{"cache"}),

		SearchQueriesTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "search_queries_total",
			Help:      "Total search queries executed",
		}, []string{"mode"}),

		SearchDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "search_duration_seconds",
			Help:      "Search query duration in seconds",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		}, []string{"mode"}),
	}
}
