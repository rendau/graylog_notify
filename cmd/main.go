package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env"
	"github.com/rendau/graylog_notify/internal/core"
	"github.com/rendau/graylog_notify/internal/destination/tgram"
	"github.com/rendau/graylog_notify/internal/httphandler"
	"github.com/rs/cors"
)

type Conf struct {
	Debug            bool   `env:"DEBUG" envDefault:"false"`
	HttpPort         string `env:"HTTP_PORT" envDefault:"80"`
	HttpCors         bool   `env:"HTTP_CORS" envDefault:"false"`
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN" required:"true"`
	TelegramChatId   int64  `env:"TELEGRAM_INIT_CHAT_ID"`
}

type App struct {
	conf        Conf
	destination core.DestinationI
	core        *core.Core

	httpServer *http.Server

	exitCode int
}

func (a *App) Init() {
	// config
	{
		if err := env.Parse(&a.conf); err != nil {
			panic(err)
		}
	}

	// logger
	{
		if !a.conf.Debug {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			slog.SetDefault(logger)
		}
	}

	// destination
	{
		if a.conf.TelegramBotToken != "" {
			var err error
			a.destination, err = tgram.NewTGram(a.conf.TelegramBotToken, a.conf.TelegramChatId)
			a.errAssert(err, "fail to tgram.NewTGram")
		}
	}

	// core
	{
		var err error
		a.core, err = core.NewCore(a.destination)
		a.errAssert(err, "fail to core.NewCore")
	}

	// http server
	{
		// router
		router := httphandler.NewHttpHandler(a.core)

		// add cors middleware
		if a.conf.HttpCors {
			router = cors.New(cors.Options{
				AllowOriginFunc: func(origin string) bool { return true },
				AllowedMethods: []string{
					http.MethodGet,
					http.MethodPut,
					http.MethodPatch,
					http.MethodPost,
					http.MethodDelete,
				},
				AllowedHeaders: []string{
					"Accept",
					"Content-Type",
					"X-Requested-With",
					"Authorization",
				},
				AllowCredentials: true,
				MaxAge:           604800,
			}).Handler(router)
		}

		// add recover middleware
		router = func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					if recErr := recover(); recErr != nil {
						slog.Error("Recovered from panic", slog.Any("err", recErr))
						w.WriteHeader(http.StatusInternalServerError)
					}
				}()
				h.ServeHTTP(w, r)
			})
		}(router)

		// server
		a.httpServer = &http.Server{
			Addr:              ":" + a.conf.HttpPort,
			Handler:           router,
			ReadHeaderTimeout: 2 * time.Second,
			ReadTimeout:       time.Minute,
			MaxHeaderBytes:    300 * 1024,
		}
	}
}

func (a *App) Start() {
	slog.Info("Start")

	// http server
	{
		go func() {
			err := a.httpServer.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				a.errAssert(err, "grpc-server stopped")
			}
		}()
		slog.Info("http-server started " + a.httpServer.Addr)
	}

	signalCtx, signalCtxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer signalCtxCancel()

	// wait signal
	<-signalCtx.Done()
}

func (a *App) Stop() {
	slog.Info("stopping")
}

func (a *App) WaitJobs() {
	slog.Info("waiting jobs")
}

func (a *App) errAssert(err error, msg string) {
	if err != nil {
		if msg != "" {
			err = fmt.Errorf("%s: %w", msg, err)
		}
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func (a *App) Exit() {
	slog.Info("exit")
	os.Exit(a.exitCode)
}

func main() {
	app := App{}
	app.Init()
	app.Start()
	app.Stop()
	app.WaitJobs()
	app.Exit()
}
