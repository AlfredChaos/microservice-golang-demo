package domain

import "time"

// User 用户领域模型
// 领域模型代表业务核心概念,不依赖于具体的技术实现
type User struct {
	ID        string    `bson:"_id,omitempty" json:"id"`           // 用户ID
	Username  string    `bson:"username" json:"username"`          // 用户名
	Email     string    `bson:"email" json:"email"`                // 邮箱
	CreatedAt time.Time `bson:"created_at" json:"created_at"`      // 创建时间
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`      // 更新时间
}

// NewUser 创建新用户
// 使用工厂函数确保创建的用户对象是有效的
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
