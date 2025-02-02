package application

import (
	"clusterMonitor/pkg/client/domain/ports"
	logger "clusterMonitor/pkg/logging/domain/ports"
	"fmt"
)

func PrtocolFactory(protocol string, logger logger.LoggerApplication) (ports.MonitoringMethod, error) {
	switch protocol {
	case "http":
		httpMonitoring := NewHttpMonitoring(logger)
		return httpMonitoring, nil
	//case "telegram":
	//	sender, err := NewTelegramSender(configSender)
	//	if err != nil || !status {
	//		panic("error creating telegram sender")
	//	}
	//	return sender, nil
	default:
		panic(fmt.Sprint("error creating %s protocol", protocol))
	}
}
