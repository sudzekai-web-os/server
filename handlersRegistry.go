package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sudzekai-web-os/core"
)

type HandlersRegistry struct {
	middlewares    map[int]core.Middleware
	handlers       map[string]http.HandlerFunc
	handlerCreator *HandlerCreator
	logger         core.ILogger
}

func NewHandlersRegistry(loggerFactory core.ILoggerFactory) core.IHandlersRegistry {
	return &HandlersRegistry{
		middlewares:    make(map[int]core.Middleware),
		handlers:       make(map[string]http.HandlerFunc),
		handlerCreator: NewHandlerCreator(loggerFactory),
		logger:         loggerFactory.NewLogger("handlers-registerer"),
	}
}

func (hr *HandlersRegistry) AddHandler(
	pattern string,
	handler core.HandlerFunc,
) core.IHandlersRegistry {
	if err := validateHandlerRegistration(pattern, handler); err != nil {
		hr.logRegistrationError(pattern, err)
		return hr
	}

	hr.handlers[pattern] = hr.handlerCreator.CreateHandler(handler)
	hr.logger.LogDebug("добавлен обработчик для %s", pattern)

	return hr
}

func (hr *HandlersRegistry) AddNoFilterHandler(
	pattern string,
	handler http.HandlerFunc,
) core.IHandlersRegistry {
	if err := validateHTTPHandler(pattern, handler); err != nil {
		hr.logRegistrationError(pattern, err)
		return hr
	}

	hr.handlers[pattern] = handler
	hr.logger.LogDebug(
		"добавлен обработчик без фильтра результата для %s",
		pattern,
	)

	return hr
}

func (hr *HandlersRegistry) AddProtectedHandler(
	pattern string,
	handler core.HandlerFunc,
	roles []string,
) core.IHandlersRegistry {
	if err := validateHandlerRegistration(pattern, handler); err != nil {
		hr.logRegistrationError(pattern, err)
		return hr
	}

	createdHandler, err := hr.handlerCreator.CreateProtectedHandler(
		handler,
		roles,
	)

	if err != nil {
		hr.logRegistrationError(pattern, err)
		return hr
	}

	hr.handlers[pattern] = createdHandler
	hr.logger.LogDebug(
		"добавлен защищенный обработчик для %s с разрешёнными ролями: %s",
		pattern,
		strings.Join(roles, ","),
	)

	return hr
}

func (hr *HandlersRegistry) AddProtectedNoFilterHandler(
	pattern string,
	handler http.HandlerFunc,
	roles []string,
) core.IHandlersRegistry {
	if err := validateHTTPHandler(pattern, handler); err != nil {
		hr.logRegistrationError(pattern, err)
		return hr
	}

	createdHandler, err := hr.handlerCreator.CreateNoFilterProtectedHandler(
		handler,
		roles,
	)

	if err != nil {
		hr.logRegistrationError(pattern, err)
		return hr
	}

	hr.handlers[pattern] = createdHandler
	hr.logger.LogDebug(
		"добавлен защищенный обработчик для %s с разрешёнными ролями: %s",
		pattern,
		strings.Join(roles, ","),
	)

	return hr
}

func (hr *HandlersRegistry) AddMiddleware(
	position int,
	middleware core.Middleware,
) core.IHandlersRegistry {
	if middleware == nil {
		hr.logger.LogError(
			"ошибка добавления промежуточного ПО: middleware равен nil",
		)
		return hr
	}

	position = normalizeMiddlewarePosition(
		position,
		len(hr.middlewares),
	)

	hr.middlewares[position] = middleware
	hr.logger.LogDebug(
		"добавлено промежуточное ПО в позиции %d",
		position,
	)

	return hr
}

func (hr *HandlersRegistry) SetJwtMiddleware(
	jwt core.JwtMiddleware,
) core.IHandlersRegistry {
	hr.handlerCreator.SetJwtMiddleware(jwt)
	hr.logger.LogDebug("установлено защищающее промежуточное ПО")

	return hr
}

func (hr *HandlersRegistry) SetResultFilter(
	filter core.ResultFilter,
) core.IHandlersRegistry {
	hr.handlerCreator.SetResultFilter(filter)
	hr.logger.LogDebug("установлен фильтр ответов")

	return hr
}

func (hr *HandlersRegistry) GetHandler() http.Handler {
	mux := http.NewServeMux()

	for pattern, handler := range hr.handlers {
		mux.HandleFunc(pattern, handler)
		hr.logger.LogDebug(
			"зарегистрирован обработчик для %s",
			pattern,
		)
	}

	var handler http.Handler = mux

	for position, middleware := range hr.middlewares {
		handler = middleware(handler)
		hr.logger.LogDebug(
			"зарегистрировано промежуточное ПО в позиции %d",
			position,
		)
	}

	return handler
}

func (hr *HandlersRegistry) GetRoutes() []string {
	routes := make([]string, 0, len(hr.handlers))

	for route := range hr.handlers {
		routes = append(routes, route)
	}

	return routes
}

func (hr *HandlersRegistry) ClearRoutes() core.IHandlersRegistry {
	clear(hr.handlers)

	return hr
}

func (hr *HandlersRegistry) logRegistrationError(
	pattern string,
	err error,
) {
	hr.logger.LogError(
		"ошибка добавления обработчика для %s: %s. обработчик пропускается...",
		pattern,
		err.Error(),
	)
}

func validateHandlerRegistration(
	pattern string,
	handler core.HandlerFunc,
) error {
	if handler == nil {
		return fmt.Errorf("handler равен nil")
	}

	return validatePattern(pattern)
}

func validateHTTPHandler(
	pattern string,
	handler http.HandlerFunc,
) error {
	if handler == nil {
		return fmt.Errorf("handler равен nil")
	}

	return validatePattern(pattern)
}

func normalizeMiddlewarePosition(
	position int,
	length int,
) int {
	if position < 0 {
		return 0
	}

	if position > length {
		return length
	}

	return position
}

func validatePattern(pattern string) error {
	fields := strings.Fields(pattern)

	if len(fields) != 2 {
		return fmt.Errorf(
			"неверный формат паттерна %s, ожидается: GET /example/{route}",
			pattern,
		)
	}

	return nil
}
