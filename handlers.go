package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func validateExpenseRequest(req CreateExpenseRequest) error {
	if req.AmountCents <= 0 {
		return errors.New("amount must be bigger than zero")
	}
	return nil
}

func buildExpense(c *gin.Context) (Expense, error) {
	var req CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return Expense{}, errors.New("invalid request body")
	}
	if err := validateExpenseRequest(req); err != nil {
		return Expense{}, err
	}
	expense := Expense{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		AmountCents: req.AmountCents,
		Category:    req.Category,
		Description: req.Description,
	}

	return expense, nil

}

func findExpenseByID(expenses []Expense, expenseID uuid.UUID) (Expense, int, error) {
	for idx, expense := range expenses {
		if expense.ID == expenseID {
			return expense, idx, nil
		}
	}
	return Expense{}, -1, errors.New("expense not found")
}

func createExpenseHandler(expenses *[]Expense) gin.HandlerFunc {
	return func(c *gin.Context) {
		expense, err := buildExpense(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		*expenses = append(*expenses, expense)
		c.JSON(http.StatusCreated, expense)
	}
}

func getExpensesHandler(expenses *[]Expense) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, *expenses)
	}
}

func getExpenseByIDHandler(expenses *[]Expense) gin.HandlerFunc {
	return func(c *gin.Context) {
		expenseID := c.Param("id")
		parsedExpenseID, err := uuid.Parse(expenseID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		expense, _, err := findExpenseByID(*expenses, parsedExpenseID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, expense)
	}
}

func deleteExpenseHandler(expenses *[]Expense) gin.HandlerFunc {
	return func(c *gin.Context) {
		expenseID := c.Param("id")
		parsedExpenseID, err := uuid.Parse(expenseID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		expense, idx, err := findExpenseByID(*expenses, parsedExpenseID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		*expenses = append((*expenses)[:idx], (*expenses)[idx+1:]...)
		c.JSON(http.StatusOK, expense)
	}
}

func updateExpenseHandler(expenses *[]Expense) gin.HandlerFunc {
	return func(c *gin.Context) {
		expenseID := c.Param("id")
		parsedExpenseID, err := uuid.Parse(expenseID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		_, idx, err := findExpenseByID(*expenses, parsedExpenseID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		var newExpense CreateExpenseRequest
		if err := c.ShouldBindJSON(&newExpense); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if err := validateExpenseRequest(newExpense); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		(*expenses)[idx].AmountCents = newExpense.AmountCents
		(*expenses)[idx].Category = newExpense.Category
		(*expenses)[idx].Description = newExpense.Description
		(*expenses)[idx].UpdatedAt = time.Now()
		c.JSON(http.StatusOK, (*expenses)[idx])

	}
}
