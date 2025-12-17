package logging

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"

	"github.com/sirupsen/logrus"
)

type Fields = logrus.Fields

type FieldLogger interface {
	logrus.FieldLogger
}

var (
	log  logrus.FieldLogger
	once sync.Once
)

// Init initialize global instance of logging library.
func Init(fields Fields) FieldLogger {
	once.Do(func() {
		logger := logrus.New()

		logger.SetFormatter(&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})

		logger.SetLevel(logrus.InfoLevel)

		log = logger.WithFields(fields)
	})
	return log
}

type ctxKey int

const (
	loggerKey ctxKey = iota
)

func AddToContext(ctx context.Context, log FieldLogger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

func FromContext(ctx context.Context) FieldLogger {
	log, ok := ctx.Value(loggerKey).(FieldLogger)
	if ok {
		return log
	}
	return log
}

func WithFields(fields Fields) FieldLogger {
	return log.WithFields(fields)
}

func WithField(key string, value any) FieldLogger {
	return log.WithField(key, value)
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Warning(args ...interface{}) {
	log.Warning(args...)
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func LogCommand(command interface{}) {
	context, err := json.Marshal(command)
	if err != nil {
		log.Errorf("Failed to marshal command for logging: %v", err)
		// This will only break the command logging so we can continue.
	}

	commandName := reflect.TypeOf(command).Elem().Name()

	var commandAsFields logrus.Fields
	if err := json.Unmarshal(context, &commandAsFields); err != nil {
		log.Errorf("Failed to unmarshal command for logging: %v", err)
		// This will only break the command logging so we can continue.
	}

	log.
		WithFields(logrus.Fields{
			"command_name": commandName,
			"command":      commandAsFields,
		}).Info("Start command " + commandName)
}

func LogCommandFailed(command interface{}, err error) {
	commandName := reflect.TypeOf(command).Elem().Name()
	log.
		WithFields(logrus.Fields{
			"reason": err.Error(),
		}).Info("Command failed" + commandName)
}
