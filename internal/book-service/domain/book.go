package domain

import "time"

// Book book领域模型
type Book struct {
	ID        string    // 用户ID
	Bookname  string    // 用户名
	Email     string    // 邮箱
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

// NewBook 创建新book
func NewBook(Bookname, email string) *Book {
	now := time.Now()
	return &Book{
		Bookname:  Bookname,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate 验证book数据
func (u *Book) Validate() error {
	if u.Bookname == "" {
		return ErrInvalidBookname
	}
	if u.Email == "" {
		return ErrInvalidEmail
	}
	return nil
}
