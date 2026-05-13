package repository

import "errors"

var (
	ErrNotFound    = errors.New("record not found")
	ErrDuplicate   = errors.New("record already exists")
	ErrForbidden   = errors.New("access forbidden")
)
