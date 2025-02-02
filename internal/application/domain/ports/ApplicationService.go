package ports

import (
	"clusterMonitor/internal/application/domain/models"
	"github.com/go-resty/resty/v2"
)

type ApplicationService interface {
	CheckRequest(req *models.RequestInfo) (*models.Response, error)
	ExecRequest(req *models.RequestInfo) (*resty.Response, error)
}
