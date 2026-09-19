package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sudzekai-web-os/core"
)

type HandlerCreator struct {
	jwtMiddleware core.JwtMiddleware
	resultFilter  core.ResultFilter

	logger core.ILogger
}

func NewHandlerCreator(loggerFactory core.ILoggerFactory) *HandlerCreator {
	return &HandlerCreator{
		logger: loggerFactory.NewLogger("handler-creator"),
	}
}

func (hc *HandlerCreator) SetJwtMiddleware(
	jwt core.JwtMiddleware,
) *HandlerCreator {
	hc.jwtMiddleware = jwt
	return hc
}

func (hc *HandlerCreator) SetResultFilter(
	filter core.ResultFilter,
) *HandlerCreator {
	hc.resultFilter = filter
	return hc
}

func (hc *HandlerCreator) CreateProtectedHandler(
	handler core.HandlerFunc,
	roles []string,
) (http.HandlerFunc, error) {
	if err := hc.validateJwtMiddleware(); err != nil {
		return nil, err
	}

	return hc.jwtMiddleware(hc.CreateHandler(handler), roles), nil
}

func (hc *HandlerCreator) CreateNoFilterProtectedHandler(
	handler http.HandlerFunc,
	roles []string,
) (http.HandlerFunc, error) {
	if err := hc.validateJwtMiddleware(); err != nil {
		return nil, err
	}

	return hc.jwtMiddleware(handler, roles), nil
}

func (hc *HandlerCreator) CreateHandler(
	handler core.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := handler(r)

		if hc.resultFilter != nil {
			hc.resultFilter(w, r, result)
			return
		}

		hc.writeResult(w, r, result)
	}
}

func (hc *HandlerCreator) writeResult(
	w http.ResponseWriter,
	r *http.Request,
	result core.HandlerResult,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.StatusCode)

	if result.Error != nil {
		hc.writeResponse(w, r, result.Error.Error())
		return
	}

	hc.writeResponse(w, r, result.Data)
}

func (hc *HandlerCreator) writeResponse(
	w http.ResponseWriter,
	r *http.Request,
	data any,
) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		hc.logger.LogError(
			"ошибка сериализации ответа на %s %s: %s",
			r.Method,
			r.URL.Path,
			err,
		)
	}
}

func (hc *HandlerCreator) validateJwtMiddleware() error {
	if hc.jwtMiddleware == nil {
		return fmt.Errorf(
			"ошибка создания обработчика: jwtMiddleware равен nil",
		)
	}

	return nil
}
