package main

import (
	"time"

	"github.com/google/uuid"
)

type Expense struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	AmountCents int64
	Description string
	CreatedAt   time.Time
}
