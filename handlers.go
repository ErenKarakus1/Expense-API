package main

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const updateExpenseQuery = `
	UPDATE expenses
	SET
		amount_cents=$1,
		category=$2,
		description=$3,
		updated_at=NOW()
	WHERE
		id=$4
	RETURNING
		id,
		amount_cents,
		category,
		description,
		created_at,
		updated_at
	`

const getExpensesQuery = `
	SELECT 
		id,
		amount_cents,
		category,
		description,
		created_at,
		updated_at
	FROM expenses
	ORDER BY created_at DESC
	`

const createExpenseQuery = `
	INSERT INTO expenses (
		id,
		amount_cents,
		category,
		description
	)
	VALUES ($1,$2,$3,$4)
	RETURNING
		id,
		amount_cents,
		category,
		description,
		created_at,
		updated_at
	`

const findExpenseByIDQuery = `
	SELECT 
		id,
		amount_cents,
		category,
		description,
		created_at,
		updated_at
	FROM expenses 
	WHERE id=$1
	`

const deleteExpenseQuery = `DELETE FROM expenses WHERE id=$1`

var ErrExpenseNotFound = errors.New("expense not found")

var ErrInternalServer = errors.New("internal server error")

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func validateExpenseRequest(req CreateExpenseRequest) error {
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
		AmountCents: req.AmountCents,
		Category:    strings.TrimSpace(req.Category),
		Description: strings.TrimSpace(req.Description),
	}

	return expense, nil

}

func findExpenseByID(ctx context.Context, pool *pgxpool.Pool, expenseID uuid.UUID) (Expense, error) {
	var expense Expense
	err := pool.QueryRow(
		ctx,
		findExpenseByIDQuery,
		expenseID,
	).Scan(
		&expense.ID,
		&expense.AmountCents,
		&expense.Category,
		&expense.Description,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Expense{}, ErrExpenseNotFound
		}
		return Expense{}, err
	}
	return expense, nil
}

func createExpenseHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		expense, err := buildExpense(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var createdExpense Expense
		err = pool.QueryRow(
			c.Request.Context(),
			createExpenseQuery,
			expense.ID,
			expense.AmountCents,
			expense.Category,
			expense.Description,
		).Scan(
			&createdExpense.ID,
			&createdExpense.AmountCents,
			&createdExpense.Category,
			&createdExpense.Description,
			&createdExpense.CreatedAt,
			&createdExpense.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
			return
		}
		c.JSON(http.StatusCreated, createdExpense)
	}
}

func getExpensesHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(
			c.Request.Context(),
			getExpensesQuery,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
			return
		}
		defer rows.Close()

		expenses := []Expense{}

		for rows.Next() {
			var expense Expense
			err := rows.Scan(
				&expense.ID,
				&expense.AmountCents,
				&expense.Category,
				&expense.Description,
				&expense.CreatedAt,
				&expense.UpdatedAt,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
				return
			}
			expenses = append(expenses, expense)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
			return
		}
		c.JSON(http.StatusOK, expenses)
	}
}

func getExpenseByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		expenseID := c.Param("id")
		parsedExpenseID, err := uuid.Parse(expenseID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		expense, err := findExpenseByID(c.Request.Context(), pool, parsedExpenseID)
		if err != nil {
			if errors.Is(err, ErrExpenseNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
			return
		}
		c.JSON(http.StatusOK, expense)
	}
}

func deleteExpenseHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		expenseID := c.Param("id")
		parsedExpenseID, err := uuid.Parse(expenseID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		tag, err := pool.Exec(
			c.Request.Context(),
			deleteExpenseQuery,
			parsedExpenseID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrExpenseNotFound.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func updateExpenseHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		expenseID := c.Param("id")
		parsedExpenseID, err := uuid.Parse(expenseID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
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
		var updatedExpense Expense
		err = pool.QueryRow(
			c.Request.Context(),
			updateExpenseQuery,
			newExpense.AmountCents,
			strings.TrimSpace(newExpense.Category),
			strings.TrimSpace(newExpense.Description),
			parsedExpenseID,
		).Scan(
			&updatedExpense.ID,
			&updatedExpense.AmountCents,
			&updatedExpense.Category,
			&updatedExpense.Description,
			&updatedExpense.CreatedAt,
			&updatedExpense.UpdatedAt,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": ErrExpenseNotFound.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
			return
		}
		c.JSON(http.StatusOK, updatedExpense)

	}
}
