package main

import (
	"time"

	"github.com/google/uuid"
)

type Expense struct {
	ID          uuid.UUID `json:"id"`
	AmountCents int64     `json:"amountcents"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdat"`
}

type CreateExpenseRequest struct {
	AmountCents int64  `json:"amountcents"`
	Category    string `json:"category"`
	Description string `json:"description"`
}
