package api

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("content not found")
	ErrForbidden    = errors.New("forbidden")
)
