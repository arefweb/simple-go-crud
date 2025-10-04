package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go-postgres-api/internal/model"
)

var ErrNotFound = errors.New("not found")

type ArticleRepository interface {
	Create(ctx context.Context, a *model.Article) error
	GetAll(ctx context.Context) ([]model.Article, error)
	Update(ctx context.Context, a *model.Article) error
}

type pgArticleRepo struct {
	pool *pgxpool.Pool
}

func NewPGArticleRepo(pool *pgxpool.Pool) ArticleRepository {
	return &pgArticleRepo{pool: pool}
}

func (r *pgArticleRepo) Create(ctx context.Context, a *model.Article) error {
	// Insert, return id and created_at
	sql := `INSERT INTO articles (title, content, author, published_at)
            VALUES ($1, $2, $3, $4)
            RETURNING id, created_at`
	var createdAt time.Time
	err := r.pool.QueryRow(ctx, sql,
		a.Title, a.Content, a.Author, a.PublishedAt).Scan(&a.ID, &createdAt)
	if err != nil {
		return err
	}
	a.CreatedAt = createdAt
	return nil
}

func (r *pgArticleRepo) GetAll(ctx context.Context) ([]model.Article, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, title, content, author, published_at, created_at FROM articles ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Article
	for rows.Next() {
		var a model.Article
		var publishedAt *time.Time
		if err := rows.Scan(&a.ID, &a.Title, &a.Content, &a.Author, &publishedAt, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.PublishedAt = publishedAt
		out = append(out, a)
	}
	return out, nil
}

func (r *pgArticleRepo) Update(ctx context.Context, a *model.Article) error {
    sql := `UPDATE articles
            SET title = $1, content = $2, author = $3, published_at = $4
            WHERE id = $5
            RETURNING id, title, content, author, published_at, created_at`
    
    err := r.pool.QueryRow(ctx, sql,
        a.Title, a.Content, a.Author, a.PublishedAt, a.ID,
    ).Scan(&a.ID, &a.Title, &a.Content, &a.Author, &a.PublishedAt, &a.CreatedAt)
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return ErrNotFound
        }
        return err
    }
    return nil
}
