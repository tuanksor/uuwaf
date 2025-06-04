package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// AuditEventsProcessed tracks the number of audit events processed
	AuditEventsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uuwaf_audit_events_processed_total",
			Help: "The total number of processed audit events",
		},
		[]string{"type"},
	)

	// PodReloads tracks the number of pod reloads
	PodReloads = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uuwaf_pod_reloads_total",
			Help: "The total number of pod reloads",
		},
		[]string{"pod", "status"},
	)

	// ProcessingDuration tracks the time taken to process events
	ProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "uuwaf_processing_duration_seconds",
			Help:    "Time taken to process audit events",
			Buckets: prometheus.DefBuckets,
		},
	)

	// LastProcessedID tracks the last processed audit ID
	LastProcessedID = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "uuwaf_last_processed_id",
			Help: "The ID of the last processed audit event",
		},
	)
) 