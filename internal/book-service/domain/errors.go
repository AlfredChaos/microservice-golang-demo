package domain

import "errors"

var (
	// ErrInvalidBookname 无效的书名
	ErrInvalidBookname = errors.New("invalid Bookname")

	// ErrInvalidEmail 无效的邮箱
	ErrInvalidEmail = errors.New("invalid email")

	// ErrBookNotFound 用户不存在
	ErrBookNotFound = errors.New("Book not found")

	// ErrBookAlreadyExists 用户已存在
	ErrBookAlreadyExists = errors.New("Book already exists")
)
