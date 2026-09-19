package server

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/sudzekai-web-os/core"
)

type Server struct {
	httpServer *http.Server
	mux        *http.ServeMux

	host string
	port int

	handlersRegistry core.IHandlersRegistry

	logger core.ILogger

	isListening bool
}

func NewServer(loggerFactory core.ILoggerFactory) core.IServer {
	return &Server{
		handlersRegistry: NewHandlersRegistry(loggerFactory),
		logger:           loggerFactory.NewLogger("server"),
	}
}

func (srv *Server) SetHost(host string) {
	srv.host = host
}

func (srv *Server) SetPort(port int) {
	srv.port = port
}

func (srv *Server) Start() {
	if srv.isListening {
		srv.logger.LogError("ошибка запуска сервера: сервер уже запущен")
	}

	srv.httpServer = &http.Server{
		Addr:    net.JoinHostPort(srv.host, strconv.Itoa(srv.port)),
		Handler: srv.handlersRegistry.GetHandler(),
	}

	srv.isListening = true

	srv.logger.LogInformation("сервер запущен и слушает " + srv.host + ":" + strconv.Itoa(srv.port))

	err := srv.httpServer.ListenAndServe()

	srv.isListening = false

	if err != nil && err != http.ErrServerClosed {
		srv.logger.LogError(
			"ошибка HTTP-сервера: %s", err,
		)
	}
}

func (srv *Server) Stop() {
	if !srv.isListening {
		srv.logger.LogError("ошибка остановки сервера: сервер не запущен")
	}

	srv.isListening = false

	if err := srv.httpServer.Close(); err != nil {
		srv.logger.LogError("ошибка остановки сервера: %s", err.Error())
	}

	srv.logger.LogInformation("Сервер остановлен")
}

func (srv *Server) WaitForShutdown() {
	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		os.Interrupt,
	)

	defer signal.Stop(signalChan)

	<-signalChan

	srv.Stop()
}

func (srv *Server) IsListening() bool {
	return srv.isListening
}

func (srv *Server) GetRegistry() core.IHandlersRegistry {
	return srv.handlersRegistry
}
