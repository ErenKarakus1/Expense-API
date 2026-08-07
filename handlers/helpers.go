package handlers

import (
	"errors"
	"strings"

	"github.com/ErenKarakus1/Expense-API/models"
	"github.com/ErenKarakus1/Expense-API/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func buildExpense(c *gin.Context) (models.Expense, error) {
	var req models.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return models.Expense{}, errors.New("invalid request body")
	}
	if err := repository.ValidateExpenseRequest(req); err != nil {
		return models.Expense{}, err
	}
	expense := models.Expense{
		ID:          uuid.New(),
		AmountCents: req.AmountCents,
		Category:    strings.TrimSpace(req.Category),
		Description: strings.TrimSpace(req.Description),
	}

	return expense, nil
}
