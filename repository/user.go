package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/ErenKarakus1/Expense-API/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailExists = errors.New("This email is already registered")

const createUserQuery = `
	INSERT INTO users (
		id,
		name,
		email,
		password_hash
	)
	VALUES ($1,$2,$3,$4)
	RETURNING
		id,
		name,
		email,
		created_at
`

const findUserQuery = `
	SELECT
		id,
		name,
		email,
		password_hash,
		created_at
	FROM users
	WHERE email=$1
`

func FindUserByEmail(ctx context.Context, pool *pgxpool.Pool, userEmail string) (models.User, error) {
	email := strings.ToLower(strings.TrimSpace(userEmail))
	var user models.User
	err := pool.QueryRow(
		ctx,
		findUserQuery,
		email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, errors.New("invalid email or password")
		}
		return models.User{}, err
	}
	return user, nil
}

func CreateUser(ctx context.Context, pool *pgxpool.Pool, user models.User) (models.CreatedUser, error) {
	var createdUser models.CreatedUser
	err := pool.QueryRow(
		ctx,
		createUserQuery,
		user.ID,
		user.Name,
		user.Email,
		user.PasswordHash,
	).Scan(
		&createdUser.ID,
		&createdUser.Name,
		&createdUser.Email,
		&createdUser.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.CreatedUser{}, ErrEmailExists
		}
		return models.CreatedUser{}, err
	}
	return createdUser, nil
}
