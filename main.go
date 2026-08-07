package main

import (
	"fmt"
	"log"

	"github.com/ErenKarakus1/Expense-API/auth"
	"github.com/ErenKarakus1/Expense-API/config"
	"github.com/ErenKarakus1/Expense-API/db"
	"github.com/ErenKarakus1/Expense-API/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to Postgres!")

	defer pool.Close()

	router := gin.Default()

	router.GET("/health", handlers.HealthHandler)

	router.POST("/register", handlers.RegisterHandler(pool))

	router.POST("/login", handlers.LoginHandler(pool, cfg))

	authGroup := router.Group("/")
	authGroup.Use(auth.AuthMiddleware(cfg.JWTSecret))

	authGroup.POST("/expenses", handlers.CreateExpenseHandler(pool))

	authGroup.GET("/expenses", handlers.GetExpensesHandler(pool))

	authGroup.GET("/expenses/:id", handlers.GetExpenseByIDHandler(pool))

	authGroup.DELETE("/expenses/:id", handlers.DeleteExpenseHandler(pool))

	authGroup.PUT("/expenses/:id", handlers.UpdateExpenseHandler(pool))

	router.Run(":8080")
}
