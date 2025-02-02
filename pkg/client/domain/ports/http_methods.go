package ports

import (
	"clusterMonitor/pkg/client/domain/models"
	"github.com/go-resty/resty/v2"
)

type HttpMethod interface {
	//CreateRequest(name string, url string, path []string, duration time.Duration, protocol string) *models.HttpRequest
	//MakeMonitoringRequest(request *models.MonitoringHttpRequest) ([]*resty.Response, error)
	MakeRequest(request *models.HttpRequest) (*resty.Response, error)
}
