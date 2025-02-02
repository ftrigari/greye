package models

import (
	modelsHttp "clusterMonitor/pkg/client/domain/models"
	"clusterMonitor/pkg/config/domain/models"
	"clusterMonitor/pkg/scheduler/application"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"strconv"
	"strings"
	"time"
)

type SchedulerApplication struct {
	application.Job
	modelsHttp.MonitoringHttpRequest
	ScheduledApplication string `json:"scheduledApplication"`
	MaxFailRequests      int    `json:"maxFailedRequests"`
	FailedRequest        int    `json:"failedRequest"`
}

func NewSchedulerApplication(ri RequestInfo) *SchedulerApplication {
	interval := ri.Interval * time.Second

	return &SchedulerApplication{
		Job: application.Job{
			Interval: interval,
			Ticker:   time.NewTicker(interval),
			Quit:     make(chan struct{}),
		},
		MonitoringHttpRequest: modelsHttp.MonitoringHttpRequest{
			Name:     ri.Name,
			Host:     ri.Host,
			Timeout:  ri.Timeout,
			Protocol: ri.Protocol,
			Path:     ri.Paths,
			Method:   ri.Method,
			Header:   ri.Headers,
			Body:     ri.Body,
			Interval: interval,
		},
	}
}

func GetPortUsed(svc *v1.Service) int {
	port := int(svc.Spec.Ports[0].Port)
	if portAnnotation := svc.Annotations["cm-port"]; portAnnotation != "" {
		portInt, err := strconv.Atoi(portAnnotation)
		if err != nil {
			port = int(portInt)
		}
	}
	return port
}

func NewSchedulerApplicationFromService(svc *v1.Service, defaultValue *models.Application) *SchedulerApplication {

	interval := defaultValue.IntervalSeconds * time.Second
	protocol := defaultValue.Protocol
	method := defaultValue.Method
	maxFailedRequests := defaultValue.MaxFailedRequests
	timeout := defaultValue.TimeoutSeconds * time.Second

	// If the service has annotations, use them to override the default values
	interval, err := time.ParseDuration(svc.Annotations["cm-intervalSeconds"] + "s")

	if err != nil {
		interval = defaultValue.IntervalSeconds * time.Second
	}

	if protocolAnnotation := svc.Annotations["cm-protocol"]; protocolAnnotation != "" {
		protocol = protocolAnnotation
	}

	if methodAnnotation := svc.Annotations["cm-method"]; methodAnnotation != "" {
		method = methodAnnotation
	}

	if maxFailedRequestsAnnotation := svc.Annotations["cm-maxFailedRequests"]; maxFailedRequestsAnnotation != "" {
		maxFailedRequests, err = strconv.Atoi(maxFailedRequestsAnnotation)
		if err != nil {
			maxFailedRequests = defaultValue.MaxFailedRequests
		}
	}

	if timeoutAnnotation := svc.Annotations["cm-timeoutSeconds"]; timeoutAnnotation != "" {
		timeout, err = time.ParseDuration(timeoutAnnotation + "s")
		if err != nil {
			timeout = defaultValue.TimeoutSeconds * time.Second
		}
	}

	stopTimeStr := svc.Annotations["cm-stop-monitoring-until"]
	stopTime := time.Now()
	if stopTimeStr != "" {
		stopTime, err = time.Parse("2006-01-02T15:04:05", stopTimeStr)
		if err == nil {
		}
	}

	paths := strings.Split(svc.Annotations["cm-paths"], "\n")
	port := GetPortUsed(svc)
	svcHost := fmt.Sprint(svc.Name, ".", svc.Namespace, ".svc.cluster.local")
	return &SchedulerApplication{
		Job: application.Job{
			Interval: interval,
			Ticker:   time.NewTicker(interval),
			Quit:     make(chan struct{}),
		},

		MonitoringHttpRequest: modelsHttp.MonitoringHttpRequest{
			Name:                svc.Name,
			Host:                svcHost,
			Timeout:             timeout,
			Protocol:            protocol,
			Path:                paths,
			Method:              method,
			Interval:            interval,
			Port:                port,
			StopMonitoringUntil: stopTime,
		},
		MaxFailRequests: maxFailedRequests,
		FailedRequest:   0,
	}
}

func GenerateJobSchedulerApplication(ri SchedulerApplication) *SchedulerApplication {
	interval := ri.MonitoringHttpRequest.Interval

	ri.Job = application.Job{
		Interval: interval,
		Ticker:   time.NewTicker(interval),
		Quit:     make(chan struct{}),
	}

	return &ri
}

func (s *SchedulerApplication) Validate() error {
	if s.MaxFailRequests == 0 {
		return errors.New("MaxFailedRequests must be greater than 0")
	}
	if s.MonitoringHttpRequest.Method == "" {
		return errors.New("Method must be provided")
	}
	if s.Protocol == "" {
		return errors.New("Protocol must be provided")
	}
	if s.MonitoringHttpRequest.Interval < 30 {
		return errors.New("IntervalSeconds must be greater than 30")
	}
	if s.Timeout < 1 {
		return errors.New("TimeoutSeconds must be greater or equal to 1 second")
	}
	return nil
}

func (s SchedulerApplication) GetSvcHostname() string {
	return s.Host
}
