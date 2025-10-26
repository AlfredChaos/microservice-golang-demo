package domain

import "errors"

var (
	// ErrInvalidUsername 无效的用户名
	ErrInvalidUsername = errors.New("invalid username")
	
	// ErrInvalidEmail 无效的邮箱
	ErrInvalidEmail = errors.New("invalid email")
	
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errors.New("user not found")
	
	// ErrUserAlreadyExists 用户已存在
	ErrUserAlreadyExists = errors.New("user already exists")
)
