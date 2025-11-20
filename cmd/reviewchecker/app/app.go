package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/6ermvH/avito-reviewchecker/cmd/reviewchecker/config"
	"github.com/6ermvH/avito-reviewchecker/cmd/reviewchecker/middleware"
	"github.com/6ermvH/avito-reviewchecker/internal/httpserver"
	"github.com/6ermvH/avito-reviewchecker/internal/repository/postgres"
	"github.com/6ermvH/avito-reviewchecker/internal/usecase"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type App struct {
	cfg    config.Config
	logger *slog.Logger
	server *http.Server
}

func New(cfg config.Config) (*App, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(cfg.Log.Level),
	}))

	router := chi.NewRouter()
	router.Use(middleware.RequestLogger(logger))

	db, err := sql.Open("pgx", cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("connect to db: %w", err)
	}

	registerRoutes(router, usecase.New(postgres.New(db), logger))

	httpSrv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.HTTP.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.HTTP.IdleTimeout) * time.Second,
	}

	return &App{
		cfg:    cfg,
		logger: logger,
		server: httpSrv,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		a.logger.Info("starting http server", "addr", a.cfg.HTTP.Addr)

		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http server: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("shutdown initiated")

		//nolint:mnd
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		//nolint:contextcheck
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}

		return nil
	case err := <-errCh:
		return err
	}
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}

func registerRoutes(r chi.Router, svc httpserver.Service) {
	r.Route("/team", func(r chi.Router) {
		r.Post("/add", httpserver.HandleTeamAdd(svc))
		r.Get("/get", httpserver.HandleTeamGet(svc))
	})

	r.Post("/users/setIsActive", httpserver.HandleSetUserActive(svc))
	r.Get("/users/getReview", httpserver.HandleGetUserReview(svc))

	r.Post("/pullRequest/create", httpserver.HandleCreatePR(svc))
	r.Post("/pullRequest/merge", httpserver.HandleMergePR(svc))
	r.Post("/pullRequest/reassign", httpserver.HandleReassignPR(svc))

	r.Route("/stats", func(r chi.Router) {
		r.Get("/reviewers", httpserver.HandleReviewerStats(svc))
		r.Get("/pullRequests", httpserver.HandlePullRequestStats(svc))
	})
}
