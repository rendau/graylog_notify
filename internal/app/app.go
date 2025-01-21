package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rendau/graylog_notify/internal/config"
	handlerHttpP "github.com/rendau/graylog_notify/internal/handler/http"
	serviceCoreP "github.com/rendau/graylog_notify/internal/service/core"
	serviceDestinationTelegramP "github.com/rendau/graylog_notify/internal/service/destination/telegram"
	serviceSourceGraylogP "github.com/rendau/graylog_notify/internal/service/source/graylog"
)

const (
	SystemServerPort = "3003"
)

type App struct {
	httpServer   *HttpServer
	systemServer *HttpServer

	exitCode int
}

func (a *App) Init() {
	var err error

	// logger
	{
		if !config.Conf.Debug {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			slog.SetDefault(logger)
		}
	}

	serviceSourceGraylog := serviceSourceGraylogP.New()

	telegramRules := map[string]*serviceDestinationTelegramP.Rule{}
	for k, v := range config.Conf.TelegramRules {
		telegramRules[k] = &serviceDestinationTelegramP.Rule{
			Name:   v.Name,
			ChatId: v.ChatId,
		}
	}
	serviceDestinationTelegram, err := serviceDestinationTelegramP.New(
		config.Conf.TelegramToken,
		telegramRules,
	)
	errCheck(err, "serviceDestinationTelegramP.New")

	serviceCore := serviceCoreP.New(
		serviceSourceGraylog,
		serviceDestinationTelegram,
	)

	// http server
	{
		mux := http.NewServeMux()

		handlerHttpP.AssignRoutes(mux, handlerHttpP.New(serviceCore))

		a.httpServer = NewHttpServer("main", &http.Server{
			Addr:              ":" + config.Conf.HttpPort,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			MaxHeaderBytes:    300 * 1024,
		}, true)
	}

	// system server
	{
		mux := http.NewServeMux()

		// healthcheck
		mux.HandleFunc("GET /healthcheck", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		a.systemServer = NewHttpServer("system", &http.Server{
			Addr:              ":" + SystemServerPort,
			Handler:           mux,
			ReadHeaderTimeout: 3 * time.Second,
			ReadTimeout:       5 * time.Second,
		}, false)
	}
}

func (a *App) Start() {
	slog.Info("Starting")

	// http server
	a.httpServer.Start()

	// system server
	a.systemServer.Start()
}

func (a *App) Listen() {
	signalCtx, signalCtxCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer signalCtxCancel()

	// wait signal
	<-signalCtx.Done()
}

func (a *App) Stop() {
	slog.Info("Shutting down...")

	// http server
	if err := a.httpServer.Stop(15 * time.Second); err != nil {
		slog.Error("httpServer.Stop", "error", err)
		a.exitCode = 1
	}

	// system server
	if err := a.systemServer.Stop(10 * time.Second); err != nil {
		slog.Error("systemServer.Stop", "error", err)
		a.exitCode = 1
	}
}

func (a *App) Exit() {
	slog.Info("Exit")

	os.Exit(a.exitCode)
}

func errCheck(err error, msg string) {
	if err != nil {
		if msg != "" {
			err = fmt.Errorf("%s: %w", msg, err)
		}
		slog.Error(err.Error())
		os.Exit(1)
	}
}
