package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Service metrics
	ServiceCPUUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "click_deploy_service_cpu_usage_percent",
			Help: "CPU usage percentage for a service",
		},
		[]string{"service_id", "service_name"},
	)

	ServiceMemoryUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "click_deploy_service_memory_usage_bytes",
			Help: "Memory usage in bytes for a service",
		},
		[]string{"service_id", "service_name"},
	)

	ServiceNetworkTrafficIn = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "click_deploy_service_network_traffic_in_bytes_total",
			Help: "Total incoming network traffic in bytes for a service",
		},
		[]string{"service_id", "service_name"},
	)

	ServiceNetworkTrafficOut = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "click_deploy_service_network_traffic_out_bytes_total",
			Help: "Total outgoing network traffic in bytes for a service",
		},
		[]string{"service_id", "service_name"},
	)

	ServiceRequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "click_deploy_service_requests_total",
			Help: "Total number of requests to a service",
		},
		[]string{"service_id", "service_name", "status_code"},
	)

	ServiceRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "click_deploy_service_request_duration_seconds",
			Help:    "Request duration in seconds for a service",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service_id", "service_name"},
	)

	ServiceErrorRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "click_deploy_service_error_rate",
			Help: "Error rate (errors per second) for a service",
		},
		[]string{"service_id", "service_name"},
	)

	// Database metrics
	DatabaseCPUUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "click_deploy_database_cpu_usage_percent",
			Help: "CPU usage percentage for a database",
		},
		[]string{"database_id", "database_name", "engine"},
	)

	DatabaseMemoryUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "click_deploy_database_memory_usage_bytes",
			Help: "Memory usage in bytes for a database",
		},
		[]string{"database_id", "database_name", "engine"},
	)

	DatabaseNetworkTrafficIn = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "click_deploy_database_network_traffic_in_bytes_total",
			Help: "Total incoming network traffic in bytes for a database",
		},
		[]string{"database_id", "database_name", "engine"},
	)

	DatabaseNetworkTrafficOut = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "click_deploy_database_network_traffic_out_bytes_total",
			Help: "Total outgoing network traffic in bytes for a database",
		},
		[]string{"database_id", "database_name", "engine"},
	)

	DatabaseConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "click_deploy_database_connections",
			Help: "Number of active connections to a database",
		},
		[]string{"database_id", "database_name", "engine"},
	)

	// Volume metrics
	VolumeUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "click_deploy_volume_usage_bytes",
			Help: "Volume usage in bytes",
		},
		[]string{"volume_id", "volume_name"},
	)

	VolumeIORead = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "click_deploy_volume_io_read_bytes_total",
			Help: "Total bytes read from volume",
		},
		[]string{"volume_id", "volume_name"},
	)

	VolumeIOWrite = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "click_deploy_volume_io_write_bytes_total",
			Help: "Total bytes written to volume",
		},
		[]string{"volume_id", "volume_name"},
	)
)

// RecordServiceMetrics records metrics for a service
func RecordServiceMetrics(serviceID, serviceName string, cpuPercent float64, memoryBytes int64, networkIn, networkOut int64, requestCount int, responseTime time.Duration, errorCount int) {
	ServiceCPUUsage.WithLabelValues(serviceID, serviceName).Set(cpuPercent)
	ServiceMemoryUsage.WithLabelValues(serviceID, serviceName).Set(float64(memoryBytes))
	ServiceNetworkTrafficIn.WithLabelValues(serviceID, serviceName).Add(float64(networkIn))
	ServiceNetworkTrafficOut.WithLabelValues(serviceID, serviceName).Add(float64(networkOut))
	ServiceRequestDuration.WithLabelValues(serviceID, serviceName).Observe(responseTime.Seconds())
	
	// Calculate error rate (errors per second over last minute)
	errorRate := float64(errorCount) / 60.0
	ServiceErrorRate.WithLabelValues(serviceID, serviceName).Set(errorRate)
}

// RecordServiceRequest records a single request metric
func RecordServiceRequest(serviceID, serviceName string, statusCode int, duration time.Duration) {
	ServiceRequestCount.WithLabelValues(serviceID, serviceName, string(rune(statusCode))).Inc()
	ServiceRequestDuration.WithLabelValues(serviceID, serviceName).Observe(duration.Seconds())
}

// RecordDatabaseMetrics records metrics for a database
func RecordDatabaseMetrics(databaseID, databaseName, engine string, cpuPercent float64, memoryBytes int64, networkIn, networkOut int64, connections int) {
	DatabaseCPUUsage.WithLabelValues(databaseID, databaseName, engine).Set(cpuPercent)
	DatabaseMemoryUsage.WithLabelValues(databaseID, databaseName, engine).Set(float64(memoryBytes))
	DatabaseNetworkTrafficIn.WithLabelValues(databaseID, databaseName, engine).Add(float64(networkIn))
	DatabaseNetworkTrafficOut.WithLabelValues(databaseID, databaseName, engine).Add(float64(networkOut))
	DatabaseConnections.WithLabelValues(databaseID, databaseName, engine).Set(float64(connections))
}

// RecordVolumeMetrics records metrics for a volume
func RecordVolumeMetrics(volumeID, volumeName string, usageBytes int64, readBytes, writeBytes int64) {
	VolumeUsage.WithLabelValues(volumeID, volumeName).Set(float64(usageBytes))
	VolumeIORead.WithLabelValues(volumeID, volumeName).Add(float64(readBytes))
	VolumeIOWrite.WithLabelValues(volumeID, volumeName).Add(float64(writeBytes))
}

