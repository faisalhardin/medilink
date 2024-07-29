package log

import (
	"github.com/faisalhardin/medilink/internal/library/common/log/logger"
)

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
