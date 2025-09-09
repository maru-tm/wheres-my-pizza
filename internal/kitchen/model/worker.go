package model

import (
	"errors"
)

var (
	ErrWorkerAlreadyOnline = errors.New("rmq already online")
	ErrAlreadyCooking      = errors.New("order already cooking")
)
