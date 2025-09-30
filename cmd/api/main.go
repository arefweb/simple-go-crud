package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "go-postgres-api/internal/handler"
    "go-postgres-api/internal/repository"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func main() {
    // config via env
    dbURL := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/simple_db?sslmode=disable")
    addr := getenv("HTTP_ADDR", ":8080")

    // logger
    zerolog.TimeFieldFormat = time.RFC3339
    logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

    // connect to db
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    pool, err := pgxpool.New(ctx, dbURL)
    if err != nil {
        logger.Fatal().Err(err).Msg("unable to connect to database")
    }
    defer pool.Close()

    // create repo and handler
    repo := repository.NewPGArticleRepo(pool)
    h := handler.NewArticleHandler(repo, logger.With().Str("component", "handler").Logger())

    r := chi.NewRouter()
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)   // chi request logger (simple)
    r.Use(middleware.Recoverer)

    r.Route("/articles", func(r chi.Router) {
        r.Get("/", h.List)
        r.Post("/", h.Create)
    })

    srv := &http.Server{
        Addr:    addr,
        Handler: r,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // graceful shutdown
    go func() {
        logger.Info().Str("addr", addr).Msg("starting server")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal().Err(err).Msg("server error")
        }
    }()

    // Wait for signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    logger.Info().Msg("shutting down server...")

    ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancelShutdown()
    if err := srv.Shutdown(ctxShutdown); err != nil {
        logger.Error().Err(err).Msg("server forced to shutdown")
    }

    logger.Info().Msg("server exited")
}

func getenv(key, fallback string) string {
    v := os.Getenv(key)
    if v == "" {
        return fallback
    }
    return v
}
