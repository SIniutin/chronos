package domain

import "errors"

var (
	ErrNotFound     = errors.New("gamification item not found")
	ErrDuplicateXP  = errors.New("xp transaction already exists")
	ErrInvalidInput = errors.New("invalid input")
)
