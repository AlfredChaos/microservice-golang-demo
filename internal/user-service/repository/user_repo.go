package repository

import (
	"context"

	"github.com/alfredchaos/demo/internal/user-service/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*domain.User, error)
}

type UserDocumentRepository interface {
	SaveDocument(ctx context.Context, userID string, document map[string]interface{}) error
	GetDocument(ctx context.Context, userID string) (map[string]interface{}, error)
	DeleteDocument(ctx context.Context, userID string) error

	// filter: MongoDB 查询条件，例如 bson.M{"username": "alice"}
	FindDocuments(ctx context.Context, filter map[string]interface{}, skip, limit int64) ([]map[string]interface{}, error)

	// fields: 要更新的字段，例如 map[string]interface{}{"email": "new@example.com"}
	UpdateDocumentFields(ctx context.Context, userID string, fields map[string]interface{}) error
}
