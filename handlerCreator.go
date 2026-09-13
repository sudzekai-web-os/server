package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sudzekai-web-os/abstractions"
	"github.com/sudzekai-web-os/types"
)

type HandlerCreator struct {
	jwtMiddleware types.JwtMiddleware
	resultFilter  types.ResultFilter

	loggerFactory abstractions.ILoggerFactory
}

func NewHandlerCreator(loggerFactory abstractions.ILoggerFactory) *HandlerCreator {
	return &HandlerCreator{
		loggerFactory: loggerFactory,
	}
}

func (hc *HandlerCreator) SetJwtMiddleware(jwt types.JwtMiddleware) *HandlerCreator {
	hc.jwtMiddleware = jwt
	return hc
}

func (hc *HandlerCreator) SetResultFilter(filter types.ResultFilter) *HandlerCreator {
	hc.resultFilter = filter
	return hc
}

func (hc *HandlerCreator) CreateProtectedHandler(hnd types.ProtectedHandlerFunc) (http.HandlerFunc, error) {
	if hc.jwtMiddleware == nil {
		return nil, fmt.Errorf("ошибка создания обработчика: jwtMiddleware равен nil")
	}

	handler := hc.jwtMiddleware(hnd.GetHandler(), hnd.GetRoles())

	return hc.CreateHandler(handler), nil
}

func (hc *HandlerCreator) CreateHandler(hnd types.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		result := hnd(r)

		if hc.resultFilter != nil {
			hc.resultFilter(w, r, result)
			return
		}

		w.Header().Set(
			"Content-Type",
			"application/json",
		)

		w.WriteHeader(result.StatusCode)

		if result.Error != nil {
			hc.writeResponse(w, r, result.Error.Error())
			return
		}

		hc.writeResponse(w, r, result.Data)
	}
}

func (hc *HandlerCreator) writeResponse(w http.ResponseWriter, r *http.Request, data any) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log := hc.loggerFactory.NewLogger(fmt.Sprintf("handler:%s", r.URL.Path))
		log.LogError(
			"ошибка сериализации ответа на %s %s: %s",
			r.Method,
			r.URL.Path,
			err,
		)
	}
}
