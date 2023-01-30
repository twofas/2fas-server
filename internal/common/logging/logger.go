package logging

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"reflect"
	"sync"
)

type Fields map[string]interface{}

var (
	customLogger       = New()
	defaultFields      = logrus.Fields{}
	defaultFieldsMutex = sync.RWMutex{}
)

func New() *logrus.Logger {
	logger := logrus.New()

	logger.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	logger.SetLevel(logrus.InfoLevel)

	return logger
}

func WithDefaultField(key, value string) *logrus.Logger {
	defaultFieldsMutex.Lock()
	defer defaultFieldsMutex.Unlock()

	defaultFields[key] = value

	return customLogger
}

func Info(args ...interface{}) {
	customLogger.WithFields(defaultFields).Info(args)
}

func Error(args ...interface{}) {
	customLogger.WithFields(defaultFields).Error(args)
}

func Warning(args ...interface{}) {
	customLogger.WithFields(defaultFields).Warning(args)
}

func Debug(args ...interface{}) {
	customLogger.WithFields(defaultFields).Debug(args)
}

func Fatal(args ...interface{}) {
	customLogger.WithFields(defaultFields).Fatal(args)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return customLogger.
		WithField(key, value).
		WithFields(defaultFields)
}

func WithFields(fields Fields) *logrus.Entry {
	return customLogger.
		WithFields(logrus.Fields(fields)).
		WithFields(defaultFields)
}

func LogCommand(command interface{}) {
	context, _ := json.Marshal(command)

	commandName := reflect.TypeOf(command).Elem().Name()

	var commandAsFields logrus.Fields
	json.Unmarshal(context, &commandAsFields)

	customLogger.
		WithFields(defaultFields).
		WithFields(logrus.Fields{
			"command_name": commandName,
			"command":      commandAsFields,
		}).Info("Start command " + commandName)
}

func LogCommandFailed(command interface{}, err error) {
	commandName := reflect.TypeOf(command).Elem().Name()

	customLogger.
		WithFields(defaultFields).
		WithFields(logrus.Fields{
			"reason": err.Error(),
		}).Info("Command failed" + commandName)
}
