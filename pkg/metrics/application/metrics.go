package application

import (
	"clusterMonitor/pkg/metrics/domain/ports"
	"clusterMonitor/pkg/type/domain/models"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	RoleType models.RoleType
}

const ()

func NewMetric(roleType models.RoleType) *Metrics {
	return &Metrics{
		RoleType: roleType,
	}
}

var _ ports.MetricPorts = (*Metrics)(nil)

func (m Metrics) Alarm(label string, value float64) {
	metricsInAlarm.WithLabelValues(string(m.RoleType), label).Set(value) // Set alarm status to 1
}

func (m Metrics) Monitoring(label string, value float64) {
	metricsUnderMonitoring.WithLabelValues(string(m.RoleType), label).Set(value) // Set alarm status to 1
}

var (
	metricsInAlarm = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "application_in_alarm",
			Help: "Indicates if an application is in alarm (1) or not (0).",
		},
		[]string{"type", "name"},
	)

	// Metric to track if an application is under monitoring
	metricsUnderMonitoring = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "application_under_monitoring",
			Help: "Indicates if an application is under monitoring (1) or not (0) or not presente.",
		},
		[]string{"type", "name"},
	)
)

func init() {
	// Register the gauge with Prometheus
	prometheus.MustRegister(metricsInAlarm)
	prometheus.MustRegister(metricsUnderMonitoring)
}
