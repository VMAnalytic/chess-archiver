package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

func NewLogger() logrus.FieldLogger {
	logrusLogger := logrus.New()
	logrusLogger.SetFormatter(&logrus.JSONFormatter{})
	logrusLogger.SetOutput(os.Stdout)

	lvl, _ := logrus.ParseLevel("debug")

	logrusLogger.SetLevel(lvl)

	return logrusLogger
}
