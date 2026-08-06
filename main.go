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

func validateExpense(req CreateExpenseRequest) error {
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
	if err := validateExpense(req); err != nil {
		return Expense{}, err
	}
	expense := Expense{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		AmountCents: req.AmountCents,
		Category:    req.Category,
		Description: req.Description,
	}

	return expense, nil

}

func main() {
	expenses := []Expense{}

	router := gin.Default()

	router.GET("/health", health)

	router.POST("/expenses", func(c *gin.Context) {
		expense, err := buildExpense(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		expenses = append(expenses, expense)
		c.JSON(http.StatusCreated, expense)
	})

	router.GET("/expenses", func(c *gin.Context) {
		c.JSON(http.StatusOK, expenses)
	})

	router.GET("/expenses/:id", func(c *gin.Context) {
		expenseID := c.Param("id")
		parsedExpenseID, err := uuid.Parse(expenseID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		for _, expense := range expenses {
			if expense.ID == parsedExpenseID {
				c.JSON(http.StatusOK, expense)
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "expense not found"})
	})

	router.Run(":8080")
}
