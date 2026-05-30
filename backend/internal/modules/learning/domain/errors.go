package domain

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrNotFound           = errors.New("learning item not found")
	ErrForbidden          = errors.New("forbidden")
	ErrSessionFinished    = errors.New("lesson session is finished")
	ErrNoChallenges       = errors.New("no challenges available")
	ErrNoCurrentChallenge = errors.New("no current challenge")
)
