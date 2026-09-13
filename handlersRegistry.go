package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sudzekai-web-os/abstractions"
	"github.com/sudzekai-web-os/types"
)

type HandlersRegistry struct {
	middlewares map[int]types.Middleware
	handlers    map[string]http.HandlerFunc

	handlerCreator *HandlerCreator
	logger         abstractions.ILogger
}

func NewHandlersRegistry(loggerFactory abstractions.ILoggerFactory) abstractions.IHandlersRegistry {
	return &HandlersRegistry{
		middlewares: make(map[int]types.Middleware),

		handlers: make(map[string]http.HandlerFunc),

		logger: loggerFactory.NewLogger("handlers-registerer"),
	}
}

func (hr *HandlersRegistry) AddHandler(pattern string, hnd func(r *http.Request) (result types.HandlerResult)) abstractions.IHandlersRegistry {
	err := validateHandlerRegistration(pattern, hnd)

	if err != nil {
		hr.logger.LogError("ошибка добавления обработчика для %s: %s. обработчик пропускается...", pattern, err.Error())
	}

	hr.handlers[pattern] = hr.handlerCreator.CreateHandler(hnd)
	hr.logger.LogDebug("добавлен обработчик для %s", pattern)
	return hr
}

func (hr *HandlersRegistry) AddProtectedHandler(pattern string, hnd func(r *http.Request) (result types.HandlerResult), roles []string) abstractions.IHandlersRegistry {
	err := validateHandlerRegistration(pattern, hnd)

	if err != nil {
		hr.logger.LogError("ошибка добавления обработчика для %s: %s. обработчик пропускается...", pattern, err.Error())
	}

	handler, err := hr.handlerCreator.CreateProtectedHandler(types.NewProtectedHandler(hnd, roles))

	if err != nil {
		hr.logger.LogError("ошибка добавления обработчика для %s: %s. обработчик пропускается...", pattern, err.Error())
	}

	hr.handlers[pattern] = handler
	hr.logger.LogDebug("добавлен защищенный обработчик для %s с разрешёнными ролями: %s", pattern, strings.Join(roles, ","))
	return hr
}

func (hr *HandlersRegistry) AddMiddleware(pos int, middleware types.Middleware) abstractions.IHandlersRegistry {
	position := len(hr.middlewares)

	if position >= pos {
		position = pos
	}

	hr.middlewares[position] = middleware

	hr.logger.LogDebug("добавлено промежуточное ПО в позиции %d", position)
	return hr
}

func (hr *HandlersRegistry) SetJwtMiddleware(jwt types.JwtMiddleware) abstractions.IHandlersRegistry {
	hr.handlerCreator.SetJwtMiddleware(jwt)
	hr.logger.LogDebug("установлено защищающее промежуточное ПО")
	return hr
}

func (hr *HandlersRegistry) SetResultFilter(filter types.ResultFilter) abstractions.IHandlersRegistry {
	hr.handlerCreator.SetResultFilter(filter)
	hr.logger.LogDebug("установлен фильтр ответов")
	return hr
}

func (hr *HandlersRegistry) GetHandler() http.Handler {
	mux := http.NewServeMux()

	h := http.Handler(mux)

	for pattern, handler := range hr.handlers {
		mux.HandleFunc(pattern, handler)
		hr.logger.LogDebug("зарегистрирован обработчик для %s", pattern)
	}

	for pos, middleware := range hr.middlewares {
		h = middleware(h)
		hr.logger.LogDebug("зарегистрировано промежуточное ПО в позиции %d", pos)
	}

	return h
}

func (hr *HandlersRegistry) GetRoutes() (result []string) {
	result = make([]string, 0)

	for route := range hr.handlers {
		result = append(result, route)
	}

	return
}

func (hr *HandlersRegistry) ClearRoutes() {
	for k := range hr.handlers {
		delete(hr.handlers, k)
	}
}

func validateHandlerRegistration(endpointPattern string, handler types.HandlerFunc) error {
	if handler == nil {
		return fmt.Errorf("handler равен nil")
	}

	fields := strings.Fields(endpointPattern)

	if len(fields) != 2 {
		return fmt.Errorf(
			"неверный формат паттерна %s, ожидается: GET /example/{route}",
			endpointPattern,
		)
	}

	return nil
}
