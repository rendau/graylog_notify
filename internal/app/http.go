package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/rs/cors"

	"github.com/rendau/graylog_notify/internal/config"
)

type HttpServer struct {
	name   string
	server *http.Server
}

func NewHttpServer(name string, server *http.Server, witMiddlewares bool) *HttpServer {
	if witMiddlewares {
		server.Handler = HttpMiddlewareCors(server.Handler)
		server.Handler = HttpMiddlewareContextWithoutCancel(server.Handler)
		server.Handler = HttpMiddlewareRecovery(server.Handler)
	}

	return &HttpServer{
		name:   name,
		server: server,
	}
}

func (s *HttpServer) Start() {
	go func() {
		err := s.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error(s.name + "-http-server stopped: " + err.Error())
			os.Exit(1)
		}
	}()

	slog.Info(s.name + "-server started " + s.server.Addr)
}

func (s *HttpServer) Stop(timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	ctx, ctxCancel := context.WithTimeout(context.Background(), timeout)
	defer ctxCancel()

	err := s.server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf(s.name+"-http-server.Shutdown: %w", err)
	}

	return nil
}

func HttpMiddlewareContextWithoutCancel(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r.WithContext(context.WithoutCancel(r.Context())))
	})
}

func HttpMiddlewareRecovery(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// use always new err instance in defer
			if err := recover(); err != nil {
				slog.Error("HTTP handler recovered from panic", slog.Any("error", err))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		handler.ServeHTTP(w, r)
	})
}

func HttpMiddlewareCors(handler http.Handler) http.Handler {
	if config.Conf.HttpCors {
		return cors.New(cors.Options{
			AllowOriginFunc: func(origin string) bool { return true },
			AllowedMethods: []string{
				http.MethodGet,
				http.MethodPut,
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
		}).Handler(handler)
	}

	return handler
}
