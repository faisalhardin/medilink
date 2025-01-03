package server

import (
	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/entity/http"
	authmodule "github.com/faisalhardin/medilink/internal/library/middlewares/auth"
)

type module struct {
	httpHandler      *http.Handlers
	middlewareModule *authmodule.Module
}

func LoadModules(cfg *config.Config, handlers *http.Handlers, authModule *authmodule.Module) *module {

	modules := &module{
		httpHandler:      handlers,
		middlewareModule: authModule,
	}

	return modules
}
