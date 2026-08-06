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
		return errors.New("Couldnt create expense, amount must be bigger than zero")
	}
	return nil
}

func buildExpense(c *gin.Context) (Expense, error) {
	var req CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return Expense{}, errors.New("Couldnt create expense, invalid request body")
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

	router.Run(":8080")
}
