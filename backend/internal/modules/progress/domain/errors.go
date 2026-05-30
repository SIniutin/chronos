package domain

import "errors"

var (
	ErrNotFound     = errors.New("progress not found")
	ErrInvalidInput = errors.New("invalid input")
)
