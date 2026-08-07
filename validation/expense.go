package validation

import (
	"errors"
	"strings"

	"github.com/ErenKarakus1/Expense-API/models"
)

func ValidateExpenseRequest(req models.CreateExpenseRequest) error {
	if req.AmountCents <= 0 {
		return errors.New("amount must be bigger than zero")
	}
	if len(strings.TrimSpace(req.Category)) > 50 {
		return errors.New("category too long")
	}
	if len(strings.TrimSpace(req.Description)) > 500 {
		return errors.New("description too long")
	}
	return nil
}
