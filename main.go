package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading environment")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	pool, err := pgxpool.New(
		context.Background(),
		databaseURL,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to Postgres!")

	defer pool.Close()

	router := gin.Default()

	router.GET("/health", health)

	router.POST("/expenses", createExpenseHandler(pool))

	router.GET("/expenses", getExpensesHandler(pool))

	router.GET("/expenses/:id", getExpenseByIDHandler(pool))

	router.DELETE("/expenses/:id", deleteExpenseHandler(pool))

	router.PUT("/expenses/:id", updateExpenseHandler(pool))

	router.Run(":8080")
}
