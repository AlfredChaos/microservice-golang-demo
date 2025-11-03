package cache

import (
	"context"

	"github.com/alfredchaos/demo/internal/book-service/domain"
)

const (
	// Redis Key 前缀
	bookCacheKeyPrefix = "book:id:"
)

type BookCache interface {
	SetBook(ctx context.Context, book *domain.Book, ttl int) error
	GetBook(ctx context.Context, bookID string) (*domain.Book, error)
	DeleteBook(ctx context.Context, bookID string) error
}
