# server

HTTP-сервер и реестр маршрутов для Sudzekai Web OS.

## Установка

```bash
go get github.com/sudzekai-web-os/server@latest
```

## Запуск сервера

```go
factory := logging.NewLoggerFactory(os.Stdout)
srv := server.NewServer(factory)

srv.SetHost("127.0.0.1")
srv.SetPort(8080)
srv.GetRegistry().AddHandler("GET /health", func(r *http.Request) types.HandlerResult {
	return types.HandlerResult{
		Data:       map[string]string{"status": "ok"},
		StatusCode: http.StatusOK,
	}
})

srv.Start()
```

`Start` блокирует текущую горутину до остановки HTTP-сервера. Для graceful shutdown можно вызвать `Stop`; `WaitForShutdown` ожидает `os.Interrupt`.

## Реестр обработчиков

- `AddHandler("GET /path", handler)` регистрирует обычный маршрут.
- `AddProtectedHandler` добавляет маршрут, которому нужны JWT middleware и роли.
- `AddMiddleware` добавляет стандартное `func(http.Handler) http.Handler` middleware.
- `SetJwtMiddleware` задает адаптер проверки JWT.
- `SetResultFilter` позволяет самостоятельно формировать HTTP-ответ из `types.HandlerResult`.

Обработчики возвращают `types.HandlerResult`. Без собственного result filter сервер сериализует `Data` как JSON, а при наличии ошибки отправляет текст ошибки.
