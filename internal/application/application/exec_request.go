package application

import (
	"clusterMonitor/internal/application/domain/models"
	"errors"
	"github.com/go-resty/resty/v2"
)

type ResponseWrapper struct {
	*models.Response
}

func (s *AppService) CheckRequest(req *models.RequestInfo) (*models.Response, error) {
	if req.Name == "" || req.Namespace == "" {
		return nil, errors.New("missing required fields")
	}

	response := &models.Response{
		Message: "Request is valid",
		Service: models.Service{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	}
	//
	//for _, path := range req.Path {
	//	response.Service.Check = append(response.Service.Check, models.Check{
	//		Path:       path,
	//		StatusCode: 200,
	//		Response:   "OK",
	//	})
	//}

	return response, nil
}

func (s *AppService) ExecRequest(req *models.RequestInfo) (*resty.Response, error) {
	//"https://httpbin.org/get"

	//request := s.http.CreateRequest(req.Name, req.Host, [req.Paths], time.Duration(req.Interval), req.Protocol)
	//get, err := s.http.Get(request)
	//if err != nil {
	//	return nil, err
	//}
	//return get, nil
	return nil, nil
}
