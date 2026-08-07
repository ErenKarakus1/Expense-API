package models

import (
	"time"

	"github.com/google/uuid"
)

type Expense struct {
	ID          uuid.UUID `json:"id"`
	AmountCents int64     `json:"amount_cents"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateExpenseRequest struct {
	AmountCents int64  `json:"amount_cents"`
	Category    string `json:"category"`
	Description string `json:"description"`
}
