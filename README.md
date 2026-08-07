# Expense API

A small Go REST API for tracking expenses per user. It uses Gin for HTTP routing, PostgreSQL for storage, JWT bearer tokens for authentication, and bcrypt for password hashing.

## Features

- User registration and login
- JWT authentication with `Authorization: Bearer <token>`
- Create, list, read, update, and delete expenses
- Expenses are scoped to the authenticated user
- Request validation for users and expenses
- PostgreSQL migrations in `migrations/`

## Tech Stack

- Go
- Gin
- PostgreSQL
- pgx
- JWT
- bcrypt

## Project Structure

```text
auth/          password helpers, JWT generation, auth middleware
config/        environment configuration
db/            PostgreSQL connection setup
handlers/      HTTP handlers
migrations/    database schema files
models/        request, response, and domain structs
repository/    database queries
validation/    request validation
main.go        app startup and route wiring
api_test.go    integration tests against localhost:8080
```

## Environment

Create a `.env` file in the project root:

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/expenses
JWT_SECRET=replace-with-a-long-random-secret
```

## Database

Create a PostgreSQL database named `expenses`, then run the migrations in order:

```text
migrations/001_init.sql
migrations/002_create_user_table.sql
migrations/003_add_user_id_to_expenses.sql
```

If you already had expense rows before adding users, the third migration may need a reset or a backfill because `user_id` is required.

## Run

```powershell
go run .
```

The API runs on:

```text
http://localhost:8080
```

## Authentication

Register a user:

```http
POST /register
Content-Type: application/json

{
  "name": "Test User",
  "email": "test@example.com",
  "password": "password123"
}
```

Login:

```http
POST /login
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "password123"
}
```

The login response returns a JWT:

```json
{
  "token": "..."
}
```

Use it on protected expense routes:

```http
Authorization: Bearer <token>
```

## Endpoints

### Public

```text
GET  /health
POST /register
POST /login
```

### Protected

```text
POST   /expenses
GET    /expenses
GET    /expenses/:id
PUT    /expenses/:id
DELETE /expenses/:id
```

## Expense Body

```json
{
  "amount_cents": 1500,
  "category": "Food",
  "description": "Pizza"
}
```

## Tests

The current API tests are integration tests. Start the API first:

```powershell
go run .
```

Then, in another terminal:

```powershell
go test ./...
```

The tests create unique users, log in, and call protected expense routes with bearer tokens.
