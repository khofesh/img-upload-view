package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/khofesh/img-upload-view/internal/config"
	"github.com/rs/zerolog"
)

type zerologWriter struct {
	logger *zerolog.Logger
}

func (w *zerologWriter) Write(p []byte) (n int, err error) {
	w.logger.Error().Msg(string(p))
	return len(p), nil
}

func Serve(app *config.Application) error {

	zw := &zerologWriter{logger: app.Logger}
	httpLogger := log.New(zw, "", 0)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.Port),
		Handler:      routes(app),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     httpLogger,
	}

	shutdownError := make(chan error)

	go func() {
		// intercept the signals
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.Logger.Info().Msg(fmt.Sprintf("shutting down server signal %s", s.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.Logger.Info().Msg(fmt.Sprintf("completing background tasks addr %s", srv.Addr))

		shutdownError <- nil
	}()

	app.Logger.Info().Msg(fmt.Sprintf("starting server addr %s env %s", srv.Addr, app.Config.Env))

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.Logger.Info().Msg(fmt.Sprintf("stopped server addr %s", srv.Addr))

	return nil
}
