package factories

import (
	"clusterMonitor/internal/application/application"
	applicationHandler "clusterMonitor/internal/application/infrastructure/handlers"
	clusterHandler "clusterMonitor/internal/cluster/infrastructure/handlers"
	clientApp "clusterMonitor/pkg/client/application"
	ports2 "clusterMonitor/pkg/client/domain/ports"
	configApp "clusterMonitor/pkg/config/application"
	configRepo "clusterMonitor/pkg/config/infrastructure/repositories"
	k8s "clusterMonitor/pkg/importProcess/application"
	loggerApp "clusterMonitor/pkg/logging/application"
	metricsApp "clusterMonitor/pkg/metrics/application"
	notificationApp "clusterMonitor/pkg/notification/application"
	"clusterMonitor/pkg/notification/domain/ports"
	"clusterMonitor/pkg/role/domain/models"
	schedulerApp "clusterMonitor/pkg/scheduler/application"
	"clusterMonitor/pkg/server"
	models2 "clusterMonitor/pkg/type/domain/models"
	valApp "clusterMonitor/pkg/validator/application"
	"fmt"
	"os"
	"regexp"
)

const MongoClientTimeout = 10

type Factory struct {
	//Variables
	configFilePath string
	role           models.Role

	importService *k8s.ImportProcessApplication
	configurator  *configApp.ConfigService
	validator     *valApp.Validator
	logger        *loggerApp.Logger
	httpClient    *clientApp.HttpApplication
	scheduler     *schedulerApp.Job
	notification  map[string]ports.Sender
	protocol      map[string]ports2.MonitoringMethod
	metricApp     *metricsApp.Metrics
	metricCluster *metricsApp.Metrics
}

func NewFactory(configFilePath string) *Factory {
	return &Factory{
		configFilePath: configFilePath,
	}
}

//func (f *Factory) InitializeLogger() *loggerApp.Logger {
//	if f.configurator == nil {
//		validator := f.InitializeValidator()
//		path := f.logFilePath
//
//		repo := loggerRepo.NewCSVFile(path)
//		app := loggerApp.NewLogger(repo, validator)
//		f.logger = app
//		return app
//	}
//	return f.logger
//}

func (f *Factory) InitializeValidator() *valApp.Validator {
	if f.validator == nil {
		app := valApp.NewValidator()
		f.validator = app
		return app
	}
	return f.validator
}

func (f *Factory) InitializeConfigurator() *configApp.ConfigService {
	if f.configurator == nil {
		validator := f.InitializeValidator()
		logger := f.InitializeLogger()
		path := f.configFilePath

		repo := configRepo.NewJSONRepository(path)
		//log := loggerApp.NewLogger()
		app := configApp.NewConfigService(repo, validator, logger)
		err := app.Config()
		if err != nil {
			panic(err)
		}
		f.configurator = app
		return app
	}
	return f.configurator
}

func (f *Factory) InitializeLogger() *loggerApp.Logger {
	if f.configurator == nil {
		logs := loggerApp.NewLogger()
		f.logger = logs
		return logs
	}
	return f.logger
}

func (f *Factory) InitializeHttpClient(logHandler *loggerApp.Logger) *clientApp.HttpApplication {
	if f.httpClient == nil {
		httpClient := clientApp.NewHttpApplication(logHandler)
		f.httpClient = httpClient
		return httpClient
	}
	return f.httpClient
}

func (f *Factory) InitializeScheduler() *schedulerApp.Job {
	if f.httpClient == nil {
		sched := schedulerApp.NewJob()
		f.scheduler = sched
		return sched
	}
	return f.scheduler
}

func (f *Factory) InitializeRole() models.Role {
	if f.role != "" {
		return f.role
	}
	configurator := f.InitializeConfigurator()
	log := f.InitializeLogger()
	config, _ := configurator.GetConfig()
	serverName := config.Server.ApplicationName
	hostname := os.Getenv("HOSTNAME")

	regexPattern := fmt.Sprintf(`^%s-0|%s:8080$`, serverName, serverName)

	r, err := regexp.MatchString(regexPattern, hostname)

	if err != nil || !r {
		nClusterMonitor := config.Server.NumberClusterMonitor
		regexPattern := fmt.Sprintf(`^%s(-([0-%d]))|%s(:808[0-3])$`, serverName, nClusterMonitor, serverName)

		r, err = regexp.MatchString(regexPattern, hostname)

		if err != nil || !r {
			fmt.Println(os.Environ())
			panic("the env variable 'HOSTNAME' must be set")
		}
		log.Info("I'm worker %s", hostname)
		var rt models.Role = "worker"
		f.role = rt
		return rt
	}
	log.Info("I'm controller %s", hostname)
	var rt models.Role = "controller"
	f.role = rt
	return rt

}

func (f *Factory) BuildAppHandlers() *applicationHandler.ApplicationHdl {
	logHandler := f.InitializeLogger()
	configurator := f.InitializeConfigurator()
	clientHandler := f.InitializeHttpClient(logHandler)
	schedulerHandler := f.InitializeScheduler()
	roleHandler := f.InitializeRole()
	importData := f.InitializeImportService()
	svcWatch := importData.GetKubernetesMonitoringObject()
	svcList := importData.GetKubernetesServices()
	notification := f.InitializeNotification()
	protocol := f.InitializeProtocol()
	metrics := f.initializeMetrics(models2.Application)
	appSchedulers := application.NewScheduler(clientHandler, configurator, roleHandler, svcWatch, svcList, logHandler, notification, protocol, metrics)
	validatorApp := f.InitializeValidator()
	appService := application.NewAppService(logHandler, clientHandler)

	//applicationHandler.NewApiExposedHdl(appScheduler)

	//appSchedulers := schedulerApp.NewScheduler()
	appHandlers := applicationHandler.NewApiExposedHdl(appService, validatorApp, logHandler, clientHandler, schedulerHandler, appSchedulers, configurator)
	return appHandlers
}

func (f *Factory) BuildClusterHandlers() *clusterHandler.ClusterHdl {
	logHandler := f.InitializeLogger()
	clientHandler := f.InitializeHttpClient(logHandler)
	networkInfo := server.NetworkInfo{}
	networkInfo.GetLocalIp()
	schedulerHandler := f.InitializeScheduler()
	clusterService := clusterHandler.NewClusterHandler(networkInfo, logHandler, clientHandler, schedulerHandler)
	return clusterService
}

func (f *Factory) InitializeImportService() *k8s.ImportProcessApplication {
	if f.importService == nil {
		//k8sRepo := k8sRepo.NewKubernetesRepository("localhost")
		configurator, _ := f.InitializeConfigurator().GetConfig()
		appname := configurator.Server.ApplicationName
		importApp := k8s.NewImportProcessApplication(appname)
		f.importService = importApp
	}

	return f.importService
}

func (f *Factory) InitializeNotification() map[string]ports.Sender {
	if f.notification != nil {
		return f.notification
	}
	configurator := f.InitializeConfigurator()
	config, err := configurator.GetConfig()
	if err != nil {
		panic(err)
	}
	notificationMap := make(map[string]ports.Sender)

	for notificationConfigName, notificationConfig := range config.Notification {
		notification, _ := notificationApp.NotificationSenderFactory(notificationConfigName, notificationConfig)
		notificationMap[notificationConfigName] = notification
		//notification.Send("titolo", "messaggio")
	}
	f.notification = notificationMap
	return f.notification
}

func (f *Factory) InitializeProtocol() map[string]ports2.MonitoringMethod {
	if f.protocol != nil {
		return f.protocol
	}
	configurator := f.InitializeConfigurator()
	log := f.InitializeLogger()
	config, err := configurator.GetConfig()
	if err != nil {
		panic(err)
	}
	protocolMap := make(map[string]ports2.MonitoringMethod)

	for _, protocolConfig := range config.Protocol {
		prot, _ := clientApp.PrtocolFactory(protocolConfig, log)
		protocolMap[protocolConfig] = prot
	}
	f.protocol = protocolMap
	return f.protocol
}

func (f *Factory) initializeMetrics(rt models2.RoleType) *metricsApp.Metrics {
	if rt == models2.Application {
		if f.metricApp != nil {
			return f.metricApp
		}
		metricApp := metricsApp.NewMetric(rt)
		f.metricApp = metricApp
		return metricApp
	} else {
		if f.metricCluster != nil {
			return f.metricCluster
		}
		metricCluster := metricsApp.NewMetric(rt)
		f.metricCluster = metricCluster
		return metricCluster
	}
}
