package ports

import (
	"clusterMonitor/pkg/client/domain/models"
)

type MonitoringMethod interface {
	//CreateRequest(name string, url string, path []string, duration time.Duration, protocol string) *models.HttpRequest
	MakeMonitoringRequest(request models.MonitoringHttpRequest) ([]interface{}, error)
	CheckResponse([]interface{}, error) bool
}
