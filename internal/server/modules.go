package server

import (
	"github.com/faisalhardin/medilink/internal/entity/http"
)

type module struct {
	httpHandler *http.Handlers
}

func LoadModules(handlers *http.Handlers) *module {
	modules := &module{
		httpHandler: handlers,
	}

	return modules
}
