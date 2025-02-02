package application

import (
	"clusterMonitor/pkg/config/domain/models"
	"clusterMonitor/pkg/config/domain/ports"
	loggerModels "clusterMonitor/pkg/logging/domain/ports"
	validator "clusterMonitor/pkg/validator/domain/ports"
)

type ConfigService struct {
	repository    ports.ConfigRepository
	configuration *models.Config
	validator     validator.ValidatorApplication
	logger        loggerModels.LoggerApplication
}

var _ ports.ConfigApplication = (*ConfigService)(nil)

func NewConfigService(
	repository ports.ConfigRepository,
	validator validator.ValidatorApplication,
	loggers loggerModels.LoggerApplication,
) *ConfigService {
	return &ConfigService{
		repository: repository,
		validator:  validator,
		logger:     loggers,
	}
}
