package application

import (
	clientHttp "clusterMonitor/pkg/client/domain/ports"
	logger "clusterMonitor/pkg/logging/domain/ports"
)

type AppService struct {
	logger logger.LoggerApplication
	http   clientHttp.HttpMethod
	Scheduler
}

func NewAppService(logger logger.LoggerApplication, httpCLient clientHttp.HttpMethod) *AppService {
	return &AppService{logger: logger, http: httpCLient}
}
