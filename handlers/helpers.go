package handlers

import (
	"errors"
	"strings"

	"github.com/ErenKarakus1/Expense-API/auth"
	"github.com/ErenKarakus1/Expense-API/models"
	"github.com/ErenKarakus1/Expense-API/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func buildExpense(c *gin.Context) (models.Expense, error) {
	var req models.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return models.Expense{}, errors.New("invalid request body")
	}
	if err := validation.ValidateExpenseRequest(req); err != nil {
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

func buildUser(c *gin.Context) (models.User, error) {
	var registerUser models.RegisterRequest
	if err := c.ShouldBindJSON(&registerUser); err != nil {
		return models.User{}, errors.New("invalid request body")
	}
	if err := validation.ValidateRegisterRequest(registerUser); err != nil {
		return models.User{}, err
	}
	passwordHash, err := auth.GeneratePasswordHash(registerUser.Password)
	if err != nil {
		return models.User{}, err
	}
	user := models.User{
		ID:           uuid.New(),
		Name:         strings.TrimSpace(registerUser.Name),
		Email:        strings.ToLower(strings.TrimSpace(registerUser.Email)),
		PasswordHash: string(passwordHash),
	}
	return user, nil
}
