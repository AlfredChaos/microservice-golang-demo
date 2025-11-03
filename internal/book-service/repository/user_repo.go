package repository

import (
	"context"

	"github.com/alfredchaos/demo/internal/book-service/domain"
)

type BookRepository interface {
	Create(ctx context.Context, book *domain.Book) error
	GetByID(ctx context.Context, id string) (*domain.Book, error)
	GetByBookname(ctx context.Context, bookname string) (*domain.Book, error)
	Update(ctx context.Context, book *domain.Book) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*domain.Book, error)
}

type BookDocumentRepository interface {
	SaveDocument(ctx context.Context, bookID string, document map[string]interface{}) error
	GetDocument(ctx context.Context, bookID string) (map[string]interface{}, error)
	DeleteDocument(ctx context.Context, bookID string) error

	// filter: MongoDB 查询条件，例如 bson.M{"bookname": "alice"}
	FindDocuments(ctx context.Context, filter map[string]interface{}, skip, limit int64) ([]map[string]interface{}, error)

	// fields: 要更新的字段，例如 map[string]interface{}{"email": "new@example.com"}
	UpdateDocumentFields(ctx context.Context, bookID string, fields map[string]interface{}) error
}
