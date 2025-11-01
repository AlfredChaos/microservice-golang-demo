package domain

import "time"

// User 用户领域模型
type User struct {
	ID        string    // 用户ID
	Username  string    // 用户名
	Email     string    // 邮箱
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

// NewUser 创建新用户
func NewUser(username, email string) *User {
	now := time.Now()
	return &User{
		Username:  username,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate 验证用户数据
func (u *User) Validate() error {
	if u.Username == "" {
		return ErrInvalidUsername
	}
	if u.Email == "" {
		return ErrInvalidEmail
	}
	return nil
}
