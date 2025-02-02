package application

import (
	"clusterMonitor/pkg/client/domain/models"
	"clusterMonitor/pkg/client/domain/ports"
	logger "clusterMonitor/pkg/logging/domain/ports"
	"errors"
	"github.com/go-resty/resty/v2"
	"strings"
	"time"
)

type HttpMonitoring struct {
	HttpClient *resty.Client
	logger     logger.LoggerApplication
}

func NewHttpMonitoring(logger logger.LoggerApplication) *HttpMonitoring {
	restyHttpClient := resty.New()
	return &HttpMonitoring{HttpClient: restyHttpClient,
		logger: logger}
}

var _ ports.MonitoringMethod = (*HttpMonitoring)(nil)

func (h HttpMonitoring) processPaths(request models.MonitoringHttpRequest, methodFunc func(string) (interface{}, error)) ([]interface{}, error) {
	var responses []interface{}

	for _, path := range request.Path {
		resp, err := methodFunc(path)
		h.LogResponse(resp, err)
		responses = append(responses, resp)
		if err != nil {
			return responses, err
		}
	}

	return responses, nil
}

// todo ho un problema con il timeout
func (h HttpMonitoring) GetMonitoring(request models.MonitoringHttpRequest) ([]interface{}, error) {
	return h.processPaths(request, func(path string) (interface{}, error) {
		va, err := h.HttpClient.SetTimeout(request.Timeout*time.Second).SetHeader("Content-Type", "application/json").
			R().Get(request.Protocol + "://" + request.Host + path)
		return va, err
	})
}

func (h HttpMonitoring) Get(request models.HttpRequest) (interface{}, error) {
	return h.HttpClient.SetTimeout(request.Timeout).
		R().Get(request.Protocol + "://" + request.Host + request.Path)
}

func (h HttpMonitoring) PostMonitoring(request models.MonitoringHttpRequest) ([]interface{}, error) {
	return h.processPaths(request, func(path string) (interface{}, error) {
		return h.HttpClient.SetTimeout(request.Timeout).
			R().
			SetBody(request.Body).
			Post(request.Protocol + "://" + request.Host + path)
	})
}

func (h HttpMonitoring) Post(request models.HttpRequest) (interface{}, error) {
	response, err := h.HttpClient.SetTimeout(request.Timeout).
		R().
		SetBody(request.Body).
		Post(request.Protocol + "://" + request.Host + request.Path)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (h HttpMonitoring) PutMonitoring(request models.MonitoringHttpRequest) ([]interface{}, error) {
	return h.processPaths(request, func(path string) (interface{}, error) {
		return h.HttpClient.SetTimeout(request.Timeout).
			R().
			SetBody(request.Body).
			Put(request.Protocol + "://" + request.Host + path)
	})
}

func (h HttpMonitoring) DeleteMonitoring(request models.MonitoringHttpRequest) ([]interface{}, error) {
	return h.processPaths(request, func(path string) (interface{}, error) {
		va, err := h.HttpClient.SetTimeout(request.Timeout*time.Second).SetHeader("Content-Type", "application/json").
			R().Delete(request.Protocol + "://" + request.Host + path)
		return va, err
	})
}

func (h HttpMonitoring) Put(request models.HttpRequest) (interface{}, error) {
	return h.HttpClient.SetTimeout(request.Timeout).
		R().
		SetBody(request.Body).
		Put(request.Protocol + "://" + request.Host + request.Path)

}

func (h HttpMonitoring) Delete(request models.HttpRequest) (interface{}, error) {
	return h.HttpClient.SetTimeout(request.Timeout).
		R().Delete(request.Protocol + "://" + request.Host + request.Path)
}

func (h HttpMonitoring) LogResponse(resp interface{}, err error) {
	if err != nil {
		h.logger.Error("Error: %s", err.Error())
		return
	}

}

func (h HttpMonitoring) MakeMonitoringRequest(r models.MonitoringHttpRequest) ([]interface{}, error) {
	switch strings.ToUpper(r.Method) {
	case "GET":
		return h.GetMonitoring(r)
	case "POST":
		return h.PostMonitoring(r)
	case "PUT":
		return h.PutMonitoring(r)
	case "DELETE":
		return h.DeleteMonitoring(r)
	}
	return nil, errors.New("Method not implemented")
}

func (h HttpMonitoring) CheckResponse(i []interface{}, err error) bool {
	isOk := true

	for _, item := range i {
		// Perform a type assertion for each item
		response, ok := item.(*resty.Response)
		if !ok {
			return false
		}
		if response.StatusCode() != 200 {
			isOk = false
		}
	}

	return isOk
}
