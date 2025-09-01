package utils

import (
	"github.com/sirupsen/logrus"
)

func NewLogger() logrus.FieldLogger {
	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.TraceLevel)
	return logger
}
