package log

import (
	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/library/common/log"
	"github.com/faisalhardin/medilink/internal/library/common/log/logger"
)

func SetupLogging(cfg *config.Config) {

	/* #nosec */
	// START -- SET UP ERROR & FATAL LOG
	errFatalLogger, err := log.NewLogger(log.Zerolog, &logger.Config{
		AppName: cfg.Server.Name,
		Level:   log.DebugLevel, // please ignore
		// LogFile:  cfg.Log.ErrorFatalFile,
		// Caller:   cfg.Log.Caller,
		UseColor: true,
		UseJSON:  true,
	})
	if err != nil {
		log.Fatal(err)
	}
	err = log.SetLogger(log.FatalLevel, errFatalLogger)
	if err != nil {
		log.Fatal(err)
	}
	// * END -- SET UP FATAL LOG

	// START -- SET UP ERROR & FATAL LOG
	errErrorLogger, err := log.NewLogger(log.Zerolog, &logger.Config{
		AppName: cfg.Server.Name,
		Level:   log.DebugLevel, // please ignore
		// LogFile:  cfg.Log.ErrorFatalFile,
		// Caller:   cfg.Log.Caller,
		UseColor: true,
		UseJSON:  true,
	})
	if err != nil {
		log.Fatal(err)
	}
	err = log.SetLogger(log.ErrorLevel, errErrorLogger)
	if err != nil {
		log.Fatal(err)
	}
	// * END -- SET UP ERROR LOG

	// * START -- SET UP DEBUG LOG
	debugLogger, _ := log.NewLogger(log.Zerolog, &logger.Config{
		AppName: cfg.Server.Name,
		Level:   log.DebugLevel, // please ignore
		// LogFile:  cfg.Log.DebugPath,
		// Caller:   cfg.Log.Caller,
		UseColor: true,
		UseJSON:  true,
	})
	if err != nil {
		log.Fatal(err)
	}
	err = log.SetLogger(log.DebugLevel, debugLogger)
	if err != nil {
		log.Fatal(err)
	}
	// * END -- SET UP DEBUG LOG
}
