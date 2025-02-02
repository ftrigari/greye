package ports

import (
	"clusterMonitor/internal/application/domain/models"
	v1 "k8s.io/api/core/v1"
)

type SchedulerService interface {
	MonitorApplication(app *models.SchedulerApplication, startupPhase bool) error
	GetApplication(url string) (map[string]models.SchedulerApplication, error)
	DeleteApplication(app *models.SchedulerApplication) error
	DeleteApplicationFromUrl(url string) error
	AddApplication(app *models.SchedulerApplication, startupPhase bool) error

	IsEnabled(svc *v1.Service) bool
}

//package po
//
//rts
//
//import (
//"clusterMonitor/internal/application/domain/models"
//v1 "k8s.io/api/core/v1"
//)
//
//type SchedulerService interface {
//	MonitorApplication(app *models.SchedulerApplication) error
//	GetApplication(url string) (map[string]models.SchedulerApplication, error)
//	DeleteApplication(url string) error
//
//	AddApplication(app *models.SchedulerApplication) error
//
//	IsEnabled(svc *v1.Service) bool
//}
