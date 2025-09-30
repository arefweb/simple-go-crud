package handler

import (
    "context"
    "encoding/json"
    "io"
    "net/http"
    "time"

    "go-postgres-api/internal/model"
    "go-postgres-api/internal/repository"
    "github.com/rs/zerolog"
)

type ArticleHandler struct {
    Repo   repository.ArticleRepository
    Logger zerolog.Logger
}

func NewArticleHandler(repo repository.ArticleRepository, logger zerolog.Logger) *ArticleHandler {
    return &ArticleHandler{Repo: repo, Logger: logger}
}

// GET /articles
func (h *ArticleHandler) List(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    // Add a small timeout so handlers are bounded
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    articles, err := h.Repo.GetAll(ctx)
    if err != nil {
        h.Logger.Error().Err(err).Msg("failed to get articles")
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    writeJSON(w, http.StatusOK, articles)
}

// POST /articles
func (h *ArticleHandler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    var payload struct {
        Title       string  `json:"title"`
        Content     string  `json:"content"`
        Author      string  `json:"author"`
        PublishedAt *string `json:"published_at,omitempty"` // ISO8601
    }
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    if err := json.Unmarshal(body, &payload); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }

    // basic validation
    if payload.Title == "" || payload.Content == "" || payload.Author == "" {
        http.Error(w, "title, content and author are required", http.StatusBadRequest)
        return
    }
    var published *time.Time
    if payload.PublishedAt != nil && *payload.PublishedAt != "" {
        t, err := time.Parse(time.RFC3339, *payload.PublishedAt)
        if err != nil {
            http.Error(w, "published_at must be RFC3339", http.StatusBadRequest)
            return
        }
        published = &t
    }

    a := &model.Article{
        Title:       payload.Title,
        Content:     payload.Content,
        Author:      payload.Author,
        PublishedAt: published,
    }

    if err := h.Repo.Create(ctx, a); err != nil {
        h.Logger.Error().Err(err).Msg("failed to create article")
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    writeJSON(w, http.StatusCreated, a)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}
