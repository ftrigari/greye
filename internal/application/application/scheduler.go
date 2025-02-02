package application

import (
	"clusterMonitor/internal/application/domain/models"
	"clusterMonitor/internal/application/domain/ports"
	modelsHttp "clusterMonitor/pkg/client/domain/models"
	clientApp "clusterMonitor/pkg/client/domain/ports"
	configPort "clusterMonitor/pkg/config/domain/ports"
	logger "clusterMonitor/pkg/logging/domain/ports"
	metricsPort "clusterMonitor/pkg/metrics/domain/ports"
	ports2 "clusterMonitor/pkg/notification/domain/ports"
	roleModel "clusterMonitor/pkg/role/domain/models"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"log"
	netUrl "net/url"
	"os"
	"regexp"
	"sync"
	"time"
)

type Scheduler struct {
	sync.RWMutex
	applications          map[string]models.SchedulerApplication
	clusterMonitorsStatus map[string]int

	http    clientApp.HttpMethod
	config  configPort.ConfigApplication
	role    roleModel.Role
	logger  logger.LoggerApplication
	alarms  map[string]ports2.Sender
	client  map[string]clientApp.MonitoringMethod
	metrics metricsPort.MetricPorts
}

func (s *Scheduler) ReadApplications() map[string]models.SchedulerApplication {
	s.RLock()
	defer s.RUnlock()
	copiedMap := make(map[string]models.SchedulerApplication)
	for k, v := range s.applications {
		copiedMap[k] = v
	}
	return copiedMap
}

func (s *Scheduler) ReadFromApplicationMap(key string) (*models.SchedulerApplication, bool) {
	s.RLock()
	defer s.RUnlock()
	app, exist := s.applications[key]
	return &app, exist
}

func (s *Scheduler) ReadFromApplicationMapByValue(key string) (models.SchedulerApplication, bool) {
	s.RLock()
	defer s.RUnlock()
	app, exist := s.applications[key]
	return app, exist
}

func (s *Scheduler) WriteToApplicationMap(key string, applications models.SchedulerApplication) {
	s.Lock()
	s.applications[key] = applications
	s.Unlock()
}

func (s *Scheduler) DeleteFromApplication(url string) {
	s.Lock()
	delete(s.applications, url)
	s.Unlock()
}

func (s *Scheduler) ReadAlarms() map[string]ports2.Sender {
	s.RLock()
	defer s.RUnlock()
	return s.alarms
}

func (s *Scheduler) ReadFromClient(key string) (clientApp.MonitoringMethod, bool) {
	s.RLock()
	defer s.RUnlock()
	app, exist := s.client[key]
	return app, exist
}

func (s *Scheduler) GetSvcHostname(app *models.SchedulerApplication) string {
	s.RLock()
	defer s.RUnlock()
	return app.GetSvcHostname()
}

func (s *Scheduler) GetHost(app *models.SchedulerApplication) string {
	s.RLock()
	defer s.RUnlock()
	return app.Host
}

func (s *Scheduler) GetScheduledApplication(app *models.SchedulerApplication) string {
	s.RLock()
	defer s.RUnlock()
	return app.ScheduledApplication
}

var _ ports.SchedulerService = (*Scheduler)(nil)

func GetApplicationInitialized(host string, http clientApp.HttpMethod, appInitialized *map[string]models.SchedulerApplication) int {
	//todo forse questo metodo va messo da un'altra parte!
	var cmstatus int
	for {
		httpRequest := modelsHttp.HttpRequest{
			Name:     host,
			Host:     host,
			Timeout:  5 * time.Second,
			Protocol: "http",
			Path:     "/api/v1/application/monitor",
			Method:   "GET",
		}

		request, err := http.MakeRequest(&httpRequest)
		if err != nil {
			log.Printf("Request failed: %v. Retrying...", err)
			time.Sleep(5 * time.Second) // Small delay before retry
			continue
		}

		if request.StatusCode() == 200 {
			cmstatus = 0
			log.Println("Received 200 response. Processing the body...")

			// Assume the body contains a map (e.g., JSON parsed as map[string]models.SchedulerApplication)
			var responseBody map[string]models.SchedulerApplication
			err = json.Unmarshal(request.Body(), &responseBody) // Ensure request.Body() is []byte
			if err != nil {
				log.Printf("Failed to parse response body: %v. Retrying...", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Add all key-value pairs to appInitialized
			for key, value := range responseBody {
				if _, exists := (*appInitialized)[key]; exists {
					panic(fmt.Sprintf("App %s already initialized", key))
				}
				cmstatus++
				(*appInitialized)[key] = value
			}

			log.Println("Received 200 response. Exiting loop.")
			break
		}

		log.Printf("Non-200 response: %d. Retrying...", request.StatusCode())
		time.Sleep(5 * time.Second)
	}
	return cmstatus
}

func (s *Scheduler) ManageStartupWorker() {

	config, err := s.config.GetConfig()
	if err != nil {
		return
	}
	appName := config.Server.ApplicationName
	var hostController string
	if appName == "localhost" {
		hostController = fmt.Sprintf("%s:8080", appName)
	} else {
		hostController = fmt.Sprintf("%s-0.%s", appName, config.Server.ServiceHAName)
	}

	s.logger.Info("hostController " + hostController)
	httpRequest := modelsHttp.HttpRequest{
		Name:     hostController,
		Host:     hostController,
		Timeout:  5 * time.Second,
		Protocol: "http",
		Path:     "/api/v1/application/monitor",
		Method:   "GET",
	}

	request, err := s.http.MakeRequest(&httpRequest)

	if request.StatusCode() == 200 {
		s.logger.Info("Received 200 response. Processing the body...")

		// Assume the body contains a map (e.g., JSON parsed as map[string]models.SchedulerApplication)
		var responseBody map[string]models.SchedulerApplication
		err = json.Unmarshal(request.Body(), &responseBody) // Ensure request.Body() is []byte
		if err != nil {
			s.logger.Info("Failed to parse response body: %v. Retrying...", err)
			time.Sleep(5 * time.Second)
			return
		}
		hostname := s.getMyHostname()
		// Add all key-value pairs to appInitialized
		for _, value := range responseBody {
			if value.ScheduledApplication == hostname {

				data := models.GenerateJobSchedulerApplication(value)

				s.MonitorApplication(data, true)
			}
		}

		s.logger.Info("Received 200 response. Exiting loop.")
	}
}

func (s *Scheduler) RemoveNoMoreUsedSvcFoundStartupPhase(svcList *v1.ServiceList, monitoredAppFromOtherPod *map[string]models.SchedulerApplication) {
	//per ogni monitoredAppFromOtherPod
	var svcMap = make(map[string]v1.Service)

	for _, svc := range svcList.Items {
		svcHost := fmt.Sprintf("%s.%s.svc.cluster.local", svc.ObjectMeta.Name, svc.ObjectMeta.Namespace)
		svcMap[svcHost] = svc
	}

	for _, app := range *monitoredAppFromOtherPod {
		if _, exists := svcMap[app.Host]; !exists { // If the service no longer exists
			s.logger.Info("Service %s no longer exists, deleting monitoring", app.Host)
			s.DeleteApplication(&app) // Remove monitoring for that app
		}
	}
}

func (s *Scheduler) ManageStartupController(svcList *v1.ServiceList, svcWatch watch.Interface, monitoredAppFromOtherPod *map[string]models.SchedulerApplication) {
	s.RemoveNoMoreUsedSvcFoundStartupPhase(svcList, monitoredAppFromOtherPod)

	//var svcToMonitor = make(map[string]v1.Service)
	var bulkMonitor = make(map[string][]*models.SchedulerApplication)

	nServicesAtStartTime := len(svcList.Items)
	servicesElaborated := 0
	go func() {
		ch := svcWatch.ResultChan()

		applicationController := make(map[string]models.SchedulerApplication)

		for event := range ch {
			servicesElaborated++
			svc := event.Object.(*v1.Service)
			metadata := svc.ObjectMeta

			s.logger.Info(fmt.Sprintf("Service %s/%s received", svc.ObjectMeta.Namespace, svc.ObjectMeta.Name))

			isEnabled := s.IsEnabled(svc)
			if isEnabled {
				s.logger.Warn(fmt.Sprintf("Adding service %s ...", svc.ObjectMeta.Name))
			}

			host := fmt.Sprintf("%s.%s.svc.cluster.local", metadata.Name, metadata.Namespace)
			appExist := false

			usedPort := models.GetPortUsed(svc)

			if servicesElaborated == nServicesAtStartTime {
				s.logger.Info("All services at startup have been elaborated, executing bulk requests")
				for bulkHostname, bulkApps := range bulkMonitor {
					s.logger.Info("Starting bulk monitoring for %s", bulkHostname)
					if r, err := regexp.MatchString("-0.|:8080", bulkHostname); err == nil && r == true {
						for _, app := range bulkApps {
							err := s.MonitorApplication(app, true)
							if err != nil {
								s.logger.Error("Error monitoring %s: %s", bulkHostname, err.Error())
							}
						}
					} else {
						ha := modelsHttp.HttpRequest{
							Name:     bulkHostname,
							Host:     bulkHostname,
							Timeout:  10 * time.Second,
							Protocol: "http",
							Path:     "/api/v1/application/monitor",
							Body:     bulkApps,
							Method:   "PUT",
						}

						for {
							request, err := s.http.MakeRequest(&ha)
							if err != nil && request.StatusCode() != 200 {
								continue
							} else {
								break
							}
						}
					}
					s.logger.Info("Bulk monitoring for %s completed", bulkHostname)
				}
				s.logger.Info("All bulk monitoring requests completed")

			}
			if _, exists := (applicationController)[host]; exists {
				appExist = true
			}

			if !isEnabled && !appExist && event.Type != watch.Deleted {
				s.logger.Warn("The application %s is not enabled and is not under monitoring. Skipping...", host)
				continue
			}

			if event.Type == watch.Deleted || (!isEnabled && appExist) {
				s.logger.Warn("The application %s is not enabled and is under monitoring. Deleting...", host)
				ReadFromApplicationMap, _ := s.ReadFromApplicationMap(host)
				err := s.DeleteApplication(ReadFromApplicationMap)
				if err != nil {
					s.logger.Error("Error during deleting the application %s.", host)
					s.logger.Error(err.Error())
					continue
				}
				delete(applicationController, host)
				//time.Sleep(1 * time.Millisecond)
				continue
			}

			config, _ := s.config.GetConfig()
			defaultValue := config.Application
			appModels := models.NewSchedulerApplicationFromService(svc, &defaultValue)

			if isEnabled && usedPort != (applicationController)[host].Port {
				s.logger.Warn("The application %s is enabled. Adding...", host)

				err := appModels.Validate()
				if err != nil {
					s.logger.Error("Error validating application: %v", err)
				}
				applicationController[host] = *appModels
				if nServicesAtStartTime <= servicesElaborated {
					s.AddApplication(appModels, true)
				} else {
					hostname := s.ChooseHostname(appModels)
					bulkMonitor[hostname] = append(bulkMonitor[hostname], appModels)
					s.WriteToApplicationMap(appModels.Host, *appModels)
				}
			}

		}

	}()
}

func (s *Scheduler) ManageStartup(svcList *v1.ServiceList, svcWatch watch.Interface, monitoredAppFromOtherPod *map[string]models.SchedulerApplication) error {
	if s.role == roleModel.Worker {
		s.ManageStartupWorker()
	} else {
		s.ManageStartupController(svcList, svcWatch, monitoredAppFromOtherPod)
	}
	return nil
}

func NewScheduler(http clientApp.HttpMethod, config configPort.ConfigApplication, roleType roleModel.Role, svcWatch watch.Interface, svcList *v1.ServiceList, logger logger.LoggerApplication, notification map[string]ports2.Sender, client map[string]clientApp.MonitoringMethod, metrics metricsPort.MetricPorts) *Scheduler {
	c, err := config.GetConfig()
	if err != nil {
		return nil
	}
	cmstatus := make(map[string]int)
	nClusterMonitor := c.Server.NumberClusterMonitor
	appName := c.Server.ApplicationName
	svcHAName := c.Server.ServiceHAName
	var monitoredAppFromOtherPod = &map[string]models.SchedulerApplication{}
	if roleType == roleModel.Controller {
		for i := 0; i < int(nClusterMonitor); i++ {
			var k string
			if appName == "localhost" {
				k = fmt.Sprintf("%s:808%d", appName, i)
			} else {
				k = fmt.Sprintf("%s-%d.%s", appName, i, svcHAName)
			}
			cmstatus[k] = 0
			if i != 0 {
				cmstatus[k] = GetApplicationInitialized(k, http, monitoredAppFromOtherPod)
			}
		}
	}
	//
	s := &Scheduler{applications: make(map[string]models.SchedulerApplication),
		http:                  http,
		config:                config,
		role:                  roleType,
		clusterMonitorsStatus: cmstatus,
		logger:                logger,
		alarms:                notification,
		client:                client,
		metrics:               metrics,
	}

	s.ManageStartup(svcList, svcWatch, monitoredAppFromOtherPod)

	return s
}

func (s *Scheduler) SendNotification(app *models.SchedulerApplication, title string, message string) {

	s.Lock()
	if app.StopMonitoringUntil.After(time.Now()) {
		s.logger.Warn("The application %s is under allarm but it has been stopped until %s", app.Host, app.StopMonitoringUntil)
		s.Unlock()
		return
	}

	app.FailedRequest = app.FailedRequest + 1
	actualFailedCount := app.FailedRequest
	maxFail := app.MaxFailRequests
	host := app.Host
	s.Unlock()

	s.logger.Error("Error application %s: %s", host, message)
	if actualFailedCount == maxFail {
		s.metrics.Alarm(host, 1)
		for _, alarm := range s.ReadAlarms() {
			//_, err := alarm.Send(title, message)
			//if err != nil {
			s.logger.Error("Failed to send alarm '%s': %v", alarm)
			//}
		}
	}

	currentTime := time.Now()
	s.logger.Error("%s Error:  %s\n", currentTime.Format("2006-01-02 3:4:5"), host)
}

func (s *Scheduler) MonitorApplication(app *models.SchedulerApplication, startupPhase bool) error {
	svcHostname := s.GetSvcHostname(app)
	application, exist := s.ReadFromApplicationMap(svcHostname)
	if !startupPhase {
		err := s.DeleteApplication(application)
		if err != nil {
			return err
		}
	}
	if !exist {
		application = app
		s.logger.Error("application not found")
	}
	s.Lock()
	if app.Quit == nil {
		app.Quit = make(chan struct{})
	}
	s.Unlock()
	s.WriteToApplicationMap(svcHostname, *app)
	s.logger.Info(fmt.Sprintf("Application added: %v\n", svcHostname))
	s.metrics.Monitoring(svcHostname, 1)
	s.metrics.Alarm(svcHostname, 0)

	go func() {
		for {
			s.Lock()
			c := app.Ticker.C
			q := app.Quit
			s.Unlock()
			select {
			case <-c:
				s.metrics.Monitoring(svcHostname, 1)

				//application := s.applications[svcHostname]
				//s.RLock()
				//s.RUnlock()

				method, exists := s.ReadFromClient(application.Protocol)
				if !exists {
					title := "Protocol undefined"
					message := fmt.Sprint("Unsupported protocol %s", application.Protocol)
					s.SendNotification(application, title, message)
					s.logger.Error(message)
					s.WriteToApplicationMap(svcHostname, *application)
					break
				}
				s.Lock()
				resp, err := method.MakeMonitoringRequest(application.MonitoringHttpRequest)
				status := method.CheckResponse(resp, err)
				s.Unlock()
				if !status {
					message := fmt.Sprintf("Application %s is unavailable.", application.Host)
					s.SendNotification(application, "Application unavailable", message)
					s.logger.Error(message)
					s.WriteToApplicationMap(svcHostname, *application)
					break
				}
				if application.FailedRequest != 0 {
					application.FailedRequest = 0
					s.WriteToApplicationMap(svcHostname, *application)

				}
				s.metrics.Alarm(svcHostname, 0)
				s.logger.Info("SUCCESS " + svcHostname)
			case <-q:
				s.metrics.Monitoring(svcHostname, 0)
				s.Lock()
				app.Ticker.Stop()
				s.Unlock()
				return
			}
		}
	}()

	return nil
}

func (s *Scheduler) GetApplication(url string) (map[string]models.SchedulerApplication, error) {
	if url != "" {
		//app, exists := s.applications[url]
		app, exists := s.ReadFromApplicationMapByValue(url)
		//app, exists := s.applications[url]
		if !exists {
			return map[string]models.SchedulerApplication{}, nil
		}

		return map[string]models.SchedulerApplication{url: app}, nil
	}

	return s.ReadApplications(), nil
}

func (s *Scheduler) DeleteApplicationFromUrl(url string) error {
	application, exists := s.ReadFromApplicationMap(url)
	if !exists {
		return nil
	}
	err := s.DeleteApplication(application)
	if err != nil {
		return err
	}
	return nil
}

func (s *Scheduler) DeleteApplication(app *models.SchedulerApplication) error {
	url := s.GetHost(app)
	if url == "" {
		return nil
	}
	hostname := s.getMyHostname()
	keyName := s.GetScheduledApplication(app)

	s.logger.Info("hostname: %s", hostname)
	s.logger.Info("keyname: %s", keyName)

	if keyName == hostname {
		if app.Quit == nil {
			s.logger.Error("The application %s is not monitored", url)
			return nil
		}
		s.logger.Info("The application is mine")
		app.Quit <- struct{}{}
	} else {
		s.logger.Info("The application is monitored by %s", keyName)
		encodedHost := netUrl.QueryEscape(url)

		deleteRequest := &modelsHttp.HttpRequest{
			Name:     keyName,
			Host:     keyName,
			Timeout:  5 * time.Second,
			Protocol: "http",
			Path:     "/api/v1/application/monitor/" + encodedHost,
			Method:   "DELETE",
		}

		_, err := s.http.MakeRequest(deleteRequest)
		if err != nil {
			s.logger.Error("Error in the request of delete application")
			s.logger.Error(err.Error())
			return err
		}
		s.logger.Info("Request deleting send to pod %s", keyName)

	}

	s.logger.Info("The application %s is deleted", url)

	s.DeleteFromApplication(url)
	return nil
}

func (s *Scheduler) ChooseHostname(app *models.SchedulerApplication) string {
	svcHostname := s.GetSvcHostname(app)
	application, _ := s.GetApplication(svcHostname)
	if len(application) != 0 {
		app.ScheduledApplication = application[svcHostname].ScheduledApplication
		return application[svcHostname].ScheduledApplication
	}

	var minKey = ""
	var minValue int
	for k, v := range s.clusterMonitorsStatus {
		if minKey == "" || minValue > v {
			minKey = k
			minValue = v
		}
	}
	minValue = minValue + 1
	app.ScheduledApplication = minKey
	s.clusterMonitorsStatus[minKey] = minValue

	return minKey
}

func (s *Scheduler) AddApplication(app *models.SchedulerApplication, startupPhase bool) error {

	hostname := s.ChooseHostname(app)
	if r, err := regexp.MatchString("-0.|:8080", hostname); err == nil && r == true {
		err := s.MonitorApplication(app, startupPhase)
		if err != nil {
			return err
		}
		return nil
	}

	ha := modelsHttp.HttpRequest{
		Name:     hostname,
		Host:     hostname,
		Timeout:  5 * time.Second,
		Protocol: "http",
		Path:     "/api/v1/application/monitor",
		Body:     []models.SchedulerApplication{*app},
		Method:   "PUT",
	}

	for {
		request, err := s.http.MakeRequest(&ha)
		if err != nil && request.StatusCode() != 200 {
			continue
		} else {
			break
		}
	}

	//s.applications[app.Host] = *app
	s.WriteToApplicationMap(app.Host, *app)
	return nil
}

func (s *Scheduler) IsEnabled(svc *v1.Service) bool {
	if svc.Annotations["cm-enabled"] == "true" {
		return true
	}
	return false
}

func (s *Scheduler) getMyHostname() string {
	config, _ := s.config.GetConfig()
	hostname := os.Getenv("HOSTNAME")

	if config.Server.ApplicationName != "localhost" {
		return fmt.Sprintf("%s.%s", hostname, config.Server.ServiceHAName)
	}
	return hostname
}
