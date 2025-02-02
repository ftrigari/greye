package handlers

import (
	"clusterMonitor/internal/application/domain/ports"
	clientHttp "clusterMonitor/pkg/client/domain/ports"
	configPort "clusterMonitor/pkg/config/domain/ports"
	logrus "clusterMonitor/pkg/logging/domain/ports"
	schedulerPort "clusterMonitor/pkg/scheduler/domain/ports"
	valPort "clusterMonitor/pkg/validator/domain/ports"
)

type ApplicationHdl struct {
	service       ports.ApplicationService
	config        configPort.ConfigApplication
	validator     valPort.ValidatorApplication
	logger        logrus.LoggerApplication
	scheduler     schedulerPort.Operation
	http          clientHttp.HttpMethod
	schedulerData ports.SchedulerService
}

var _ ports.ApiExposed = (*ApplicationHdl)(nil)

func NewApiExposedHdl(svc ports.ApplicationService, validator valPort.ValidatorApplication, logger logrus.LoggerApplication, httpCLient clientHttp.HttpMethod,
	schedulerHandler schedulerPort.Operation, schedulerApp ports.SchedulerService, config configPort.ConfigApplication) *ApplicationHdl {

	return &ApplicationHdl{
		service:       svc,
		validator:     validator,
		logger:        logger,
		http:          httpCLient,
		scheduler:     schedulerHandler,
		schedulerData: schedulerApp,
		config:        config,
	}
}
