package logging

import (
	"encoding/json"
	"reflect"
	"sync"

	"github.com/sirupsen/logrus"
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
	defaultFieldsMutex.Lock()
	defer defaultFieldsMutex.Unlock()

	customLogger.WithFields(defaultFields).Info(args...)
}

func Infof(format string, args ...interface{}) {
	defaultFieldsMutex.Lock()
	defer defaultFieldsMutex.Unlock()

	customLogger.WithFields(defaultFields).Infof(format, args...)
}

func Error(args ...interface{}) {
	defaultFieldsMutex.Lock()
	defer defaultFieldsMutex.Unlock()

	customLogger.WithFields(defaultFields).Error(args...)
}

func Warning(args ...interface{}) {
	defaultFieldsMutex.Lock()
	defer defaultFieldsMutex.Unlock()

	customLogger.WithFields(defaultFields).Warning(args...)
}

func Fatal(args ...interface{}) {
	defaultFieldsMutex.Lock()
	defer defaultFieldsMutex.Unlock()

	customLogger.WithFields(defaultFields).Fatal(args...)
}

func WithField(key string, value interface{}) *logrus.Entry {
	defaultFieldsMutex.Lock()
	defer defaultFieldsMutex.Unlock()

	return customLogger.
		WithFields(defaultFields).
		WithField(key, value)
}

func WithFields(fields Fields) *logrus.Entry {
	defaultFieldsMutex.Lock()
	defer defaultFieldsMutex.Unlock()

	return customLogger.
		WithFields(logrus.Fields(fields)).
		WithFields(defaultFields)
}

func LogCommand(command interface{}) {
	defaultFieldsMutex.Lock()
	defer defaultFieldsMutex.Unlock()

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
	defaultFieldsMutex.Lock()
	defer defaultFieldsMutex.Unlock()

	commandName := reflect.TypeOf(command).Elem().Name()

	customLogger.
		WithFields(defaultFields).
		WithFields(logrus.Fields{
			"reason": err.Error(),
		}).Info("Command failed" + commandName)
}
