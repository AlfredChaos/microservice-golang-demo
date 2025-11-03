package psql

import (
	"context"
	"fmt"
	"time"

	"github.com/alfredchaos/demo/internal/book-service/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BookPgPO Book持久化对象（PostgreSQL）
// 负责与PostgreSQL交互的数据结构
type BookPgPO struct {
	ID        string    `gorm:"column:id;primaryKey"`
	Bookname  string    `gorm:"column:Bookname;uniqueIndex;not null"`
	Email     string    `gorm:"column:email;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName 指定表名
func (BookPgPO) TableName() string {
	return "Books"
}

// BeforeCreate GORM 钩子：创建前自动设置时间戳
func (po *BookPgPO) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if po.CreatedAt.IsZero() {
		po.CreatedAt = now
	}
	if po.UpdatedAt.IsZero() {
		po.UpdatedAt = now
	}
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动刷新 UpdatedAt
func (po *BookPgPO) BeforeUpdate(tx *gorm.DB) error {
	po.UpdatedAt = time.Now()
	return nil
}

// ToDomain 将持久化对象转换为领域对象
func (po *BookPgPO) ToDomain() *domain.Book {
	return &domain.Book{
		ID:        po.ID,
		Bookname:  po.Bookname,
		Email:     po.Email,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

// FromDomainBook 从领域对象创建持久化对象
func FromDomainBook(Book *domain.Book) *BookPgPO {
	return &BookPgPO{
		ID:        Book.ID,
		Bookname:  Book.Bookname,
		Email:     Book.Email,
		CreatedAt: Book.CreatedAt,
		UpdatedAt: Book.UpdatedAt,
	}
}

// BookPgRepository PostgreSQL仓库实现
type BookPgRepository struct {
	db *gorm.DB
}

// NewBookPgRepository 创建PostgreSQL Book仓库
func NewBookPgRepository(db *gorm.DB) *BookPgRepository {
	return &BookPgRepository{db: db}
}

// Create 创建Book
func (r *BookPgRepository) Create(ctx context.Context, Book *domain.Book) error {
	// 生成UUID作为ID
	if Book.ID == "" {
		Book.ID = uuid.New().String()
	}

	// 验证Book数据
	if err := Book.Validate(); err != nil {
		return fmt.Errorf("invalid Book data: %w", err)
	}

	po := FromDomainBook(Book)
	// GORM 会自动设置 CreatedAt 和 UpdatedAt
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return fmt.Errorf("failed to create Book: %w", err)
	}

	// 将 GORM 自动生成的时间戳同步回领域对象
	Book.CreatedAt = po.CreatedAt
	Book.UpdatedAt = po.UpdatedAt

	return nil
}

// GetByID 根据ID获取Book
func (r *BookPgRepository) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	var po BookPgPO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrBookNotFound
		}
		return nil, fmt.Errorf("failed to get Book by id: %w", err)
	}
	return po.ToDomain(), nil
}

// GetByBookname 根据书名获取Book
func (r *BookPgRepository) GetByBookname(ctx context.Context, bookname string) (*domain.Book, error) {
	var po BookPgPO
	err := r.db.WithContext(ctx).Where("bookname = ?", bookname).First(&po).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrBookNotFound
		}
		return nil, fmt.Errorf("failed to get Book by Bookname: %w", err)
	}
	return po.ToDomain(), nil
}

// Update 更新Book
func (r *BookPgRepository) Update(ctx context.Context, book *domain.Book) error {
	if book.ID == "" {
		return fmt.Errorf("book id is required for update")
	}

	// 验证Book数据
	if err := book.Validate(); err != nil {
		return fmt.Errorf("invalid book data: %w", err)
	}

	po := FromDomainBook(book)
	result := r.db.WithContext(ctx).
		Model(&BookPgPO{}).
		Where("id = ?", book.ID).
		Select("bookname", "email", "updated_at").
		Updates(po)

	if result.Error != nil {
		return fmt.Errorf("failed to update Book: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrBookNotFound
	}

	// 同步 Hook 设置的时间戳到领域对象
	book.UpdatedAt = po.UpdatedAt

	return nil
}

// Delete 删除Book
func (r *BookPgRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("Book id is required for delete")
	}

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&BookPgPO{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete Book: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrBookNotFound
	}

	return nil
}

// List 列出Book
func (r *BookPgRepository) List(ctx context.Context, offset, limit int) ([]*domain.Book, error) {
	var pos []BookPgPO

	query := r.db.WithContext(ctx)

	// 设置分页参数
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	// 按创建时间倒序排列
	if err := query.Order("created_at DESC").Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to list Books: %w", err)
	}

	// 转换为领域对象
	books := make([]*domain.Book, 0, len(pos))
	for _, po := range pos {
		books = append(books, po.ToDomain())
	}

	return books, nil
}
