package model

import (
	"errors"
)

var (
	ErrWorkerAlreadyOnline = errors.New("worker already online")
	ErrAlreadyCooking      = errors.New("order already cooking")
)
