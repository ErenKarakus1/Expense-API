package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/ErenKarakus1/Expense-API/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrExpenseNotFound = errors.New("expense not found")

var ErrInternalServer = errors.New("internal server error")

const updateExpenseQuery = `
	UPDATE expenses
	SET
		amount_cents=$1,
		category=$2,
		description=$3,
		updated_at=NOW()
	WHERE
		id=$4
	AND
		user_id=$5
	RETURNING
		id,
		user_id,
		amount_cents,
		category,
		description,
		created_at,
		updated_at
	`

const getExpensesQuery = `
	SELECT 
		id,
		user_id,
		amount_cents,
		category,
		description,
		created_at,
		updated_at
	FROM expenses
	WHERE
		user_id=$1
	ORDER BY created_at DESC
	`

const createExpenseQuery = `
	INSERT INTO expenses (
		id,
		amount_cents,
		category,
		description,
		user_id
	)
	VALUES ($1,$2,$3,$4,$5)
	RETURNING
		id,
		user_id,
		amount_cents,
		category,
		description,
		created_at,
		updated_at
	`

const findExpenseByIDQuery = `
	SELECT 
		id,
		user_id,
		amount_cents,
		category,
		description,
		created_at,
		updated_at
	FROM expenses 
	WHERE id=$1
	AND user_id=$2
	`

const deleteExpenseQuery = `
	DELETE FROM expenses 
	WHERE id=$1 
	AND user_id=$2
	`

func FindExpenseByID(ctx context.Context, pool *pgxpool.Pool, expenseID uuid.UUID, userID uuid.UUID) (models.Expense, error) {
	var expense models.Expense
	err := pool.QueryRow(
		ctx,
		findExpenseByIDQuery,
		expenseID,
		userID,
	).Scan(
		&expense.ID,
		&expense.UserID,
		&expense.AmountCents,
		&expense.Category,
		&expense.Description,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Expense{}, ErrExpenseNotFound
		}
		return models.Expense{}, err
	}
	return expense, nil
}

func CreateExpense(ctx context.Context, pool *pgxpool.Pool, expense models.Expense, userID uuid.UUID) (models.Expense, error) {
	var createdExpense models.Expense
	err := pool.QueryRow(
		ctx,
		createExpenseQuery,
		expense.ID,
		expense.AmountCents,
		expense.Category,
		expense.Description,
		userID,
	).Scan(
		&createdExpense.ID,
		&createdExpense.UserID,
		&createdExpense.AmountCents,
		&createdExpense.Category,
		&createdExpense.Description,
		&createdExpense.CreatedAt,
		&createdExpense.UpdatedAt,
	)
	if err != nil {
		return models.Expense{}, err
	}
	return createdExpense, nil
}

func GetExpenses(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]models.Expense, error) {
	rows, err := pool.Query(
		ctx,
		getExpensesQuery,
		userID,
	)
	if err != nil {
		return []models.Expense{}, ErrInternalServer
	}
	defer rows.Close()

	expenses := []models.Expense{}

	for rows.Next() {
		var expense models.Expense
		err := rows.Scan(
			&expense.ID,
			&expense.UserID,
			&expense.AmountCents,
			&expense.Category,
			&expense.Description,
			&expense.CreatedAt,
			&expense.UpdatedAt,
		)
		if err != nil {
			return []models.Expense{}, ErrInternalServer
		}
		expenses = append(expenses, expense)
	}
	if err := rows.Err(); err != nil {
		return []models.Expense{}, ErrInternalServer
	}
	return expenses, nil
}

func DeleteExpense(ctx context.Context, pool *pgxpool.Pool, expenseID uuid.UUID, userID uuid.UUID) error {
	tag, err := pool.Exec(
		ctx,
		deleteExpenseQuery,
		expenseID,
		userID,
	)
	if err != nil {
		return ErrInternalServer
	}
	if tag.RowsAffected() == 0 {
		return ErrExpenseNotFound
	}
	return nil
}

func UpdateExpense(ctx context.Context, pool *pgxpool.Pool, expenseID uuid.UUID, newExpense models.CreateExpenseRequest, userID uuid.UUID) (models.Expense, error) {
	var updatedExpense models.Expense
	err := pool.QueryRow(
		ctx,
		updateExpenseQuery,
		newExpense.AmountCents,
		strings.TrimSpace(newExpense.Category),
		strings.TrimSpace(newExpense.Description),
		expenseID,
		userID,
	).Scan(
		&updatedExpense.ID,
		&updatedExpense.UserID,
		&updatedExpense.AmountCents,
		&updatedExpense.Category,
		&updatedExpense.Description,
		&updatedExpense.CreatedAt,
		&updatedExpense.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Expense{}, ErrExpenseNotFound
		}
		return models.Expense{}, ErrInternalServer
	}
	return updatedExpense, nil
}
