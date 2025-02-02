package models

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// internal/application/domain.models/myRequest.go

type RequestInfo struct {
	Name         string        `json:"name"`
	Namespace    string        `json:"namespace"`
	Host         string        `json:"host"`
	Port         string        `json:"port"`
	Protocol     string        `json:"protocol"`
	Architecture string        `json:"architecture"`
	Interval     time.Duration `json:"interval"`
	Timeout      time.Duration `json:"timeout"`
	Paths        []string      `json:"paths"`
	Method       string        `json:"method"`
	Headers      http.Header   `json:"headers"`
	Body         interface{}   `json:"body"`
}

func (s *RequestInfo) Validate() error {

	if s.Host == "" {
		if s.Name == "" && s.Namespace == "" {
			return errors.New("host or name and namespace are required")
		}
		s.Host = fmt.Sprintf("%s.%s", s.Name, s.Namespace)
	}
	if s.Port != "" {
		s.Host = fmt.Sprintf("%s:%s", s.Host, s.Port)
	}

	if s.Protocol == "" {
		s.Protocol = "http"
	}
	if s.Interval <= 10 {
		s.Interval = 10
	}
	return nil

}
