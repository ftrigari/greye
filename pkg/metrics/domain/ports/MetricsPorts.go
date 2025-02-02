package ports

type MetricPorts interface {
	Alarm(label string, value float64)
	Monitoring(label string, value float64)
}
