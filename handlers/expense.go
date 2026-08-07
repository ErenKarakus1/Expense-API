package handlers

import (
	"errors"
	"net/http"

	"github.com/ErenKarakus1/Expense-API/models"
	"github.com/ErenKarakus1/Expense-API/repository"
	"github.com/ErenKarakus1/Expense-API/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateExpenseHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uuid.UUID)
		expense, err := buildExpense(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		createdExpense, err := repository.CreateExpense(c.Request.Context(), pool, expense, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, createdExpense)
	}
}

func GetExpensesHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uuid.UUID)
		expenses, err := repository.GetExpenses(c.Request.Context(), pool, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, expenses)
	}
}

func GetExpenseByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uuid.UUID)
		expenseID := c.Param("id")
		parsedExpenseID, err := uuid.Parse(expenseID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		expense, err := repository.FindExpenseByID(c.Request.Context(), pool, parsedExpenseID, userID)
		if err != nil {
			if errors.Is(err, repository.ErrExpenseNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, expense)
	}
}

func DeleteExpenseHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uuid.UUID)
		expenseID := c.Param("id")
		parsedExpenseID, err := uuid.Parse(expenseID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		err = repository.DeleteExpense(c.Request.Context(), pool, parsedExpenseID, userID)
		if err != nil {
			if errors.Is(err, repository.ErrExpenseNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func UpdateExpenseHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uuid.UUID)
		expenseID := c.Param("id")
		parsedExpenseID, err := uuid.Parse(expenseID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var newExpense models.CreateExpenseRequest
		if err := c.ShouldBindJSON(&newExpense); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if err := validation.ValidateExpenseRequest(newExpense); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updatedExpense, err := repository.UpdateExpense(c.Request.Context(), pool, parsedExpenseID, newExpense, userID)
		if err != nil {
			if errors.Is(err, repository.ErrExpenseNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, updatedExpense)
	}
}
