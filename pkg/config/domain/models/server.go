package models

import (
	"clusterMonitor/pkg/validator/domain/ports"
	"errors"
)

type Server struct {
	Port                 int    `json:"port"`
	TlsPort              int    `json:"tlsPort"`
	NumberClusterMonitor int    `json:"numberClusterMonitor"`
	ApplicationUrl       string `json:"applicationUrl"`
	ApplicationName      string `json:"applicationName"`
	ServiceHAName        string `json:"serviceHAName"`
	ServerUrl            string `json:"serverUrl"`
}

var _ ports.Evaluable = (*Server)(nil)

func (s *Server) Validate() error {
	if s.Port == 0 {
		return errors.New("the port is required")
	}
	return nil
}
