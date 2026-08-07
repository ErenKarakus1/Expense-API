package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ErenKarakus1/Expense-API/db"
	"github.com/ErenKarakus1/Expense-API/handlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading environment")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	pool, err := db.NewPool(databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to Postgres!")

	defer pool.Close()

	router := gin.Default()

	router.GET("/health", handlers.HealthHandler)

	router.POST("/expenses", handlers.CreateExpenseHandler(pool))

	router.GET("/expenses", handlers.GetExpensesHandler(pool))

	router.GET("/expenses/:id", handlers.GetExpenseByIDHandler(pool))

	router.DELETE("/expenses/:id", handlers.DeleteExpenseHandler(pool))

	router.PUT("/expenses/:id", handlers.UpdateExpenseHandler(pool))

	router.Run(":8080")
}
