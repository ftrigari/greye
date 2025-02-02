package application

import (
	"clusterMonitor/pkg/logging/domain/ports"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

type Logger struct {
	logrus *logrus.Logger
}

var _ ports.LoggerApplication = (*Logger)(nil)

func NewLogger() *Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
	return &Logger{logrus: log}
}

func (l Logger) Info(msg string, args ...interface{}) {
	if len(args) > 0 {
		message := fmt.Sprintf(msg, args...)
		l.logrus.Info(message)
	} else {
		l.logrus.Info(msg)
	}
}

func (l Logger) Warn(msg string, args ...interface{}) {
	if len(args) > 0 {
		message := fmt.Sprintf(msg, args...)
		l.logrus.Warn(message)
	} else {
		l.logrus.Warn(msg)
	}
}

func (l Logger) Error(msg string, args ...interface{}) {
	if len(args) > 0 {
		message := fmt.Sprintf(msg, args...)
		l.logrus.Error(message)
	} else {
		l.logrus.Error(msg)
	}
}
