package log

import (
	"github.com/faisalhardin/medilink/internal/library/common/log/logger"
)

// testLogHook records WithFields logs. Tests only; not safe with t.Parallel.
var testLogHook func(level, msg string, fields logger.KV)

// SetTestLogHook installs fn for InfoWithFields and ErrorWithFields. The returned func restores the previous hook.
func SetTestLogHook(fn func(level, msg string, fields logger.KV)) func() {
	prev := testLogHook
	testLogHook = fn
	return func() { testLogHook = prev }
}

func Error(args ...interface{}) {
	errorLogger.Error(args...)
}

func Errorln(args ...interface{}) {
	errorLogger.Errorln(args...)
}

func Errorf(format string, v ...interface{}) {
	errorLogger.Errorf(format, v...)
}

func ErrorWithFields(msg string, fields logger.KV) {
	if testLogHook != nil {
		testLogHook("error", msg, fields)
	}
	errorLogger.ErrorWithFields(msg, fields)
}

func Errors(err error) {
	errorLogger.Errors(err)
}

func Fatal(args ...interface{}) {
	fatalLogger.Fatal(args...)
}

func Fatalln(args ...interface{}) {
	fatalLogger.Fatalln(args...)
}

func Fatalf(format string, v ...interface{}) {
	fatalLogger.Fatalf(format, v...)
}

func FatalWithFields(msg string, fields logger.KV) {
	fatalLogger.FatalWithFields(msg, fields)
}

func Info(args ...interface{}) {
	infoLogger.Info(args...)
}

func Infoln(args ...interface{}) {
	infoLogger.Infoln(args...)
}

func Infof(format string, v ...interface{}) {
	infoLogger.Infof(format, v...)
}

func InfoWithFields(msg string, fields logger.KV) {
	if testLogHook != nil {
		testLogHook("info", msg, fields)
	}
	infoLogger.InfoWithFields(msg, fields)
}
