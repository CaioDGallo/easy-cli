package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var defaultLogger *logrus.Logger

func init() {
	defaultLogger = logrus.New()
	defaultLogger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	defaultLogger.SetOutput(os.Stdout)
	defaultLogger.SetLevel(logrus.InfoLevel)
}

func New() *logrus.Logger {
	return defaultLogger
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return defaultLogger.WithFields(fields)
}

func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func SetOutput(output io.Writer) {
	defaultLogger.SetOutput(output)
}

func SetLevel(level logrus.Level) {
	defaultLogger.SetLevel(level)
}
