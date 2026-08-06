package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	expenses := []Expense{}

	router := gin.Default()

	router.GET("/health", health)

	router.POST("/expenses", createExpenseHandler(&expenses))

	router.GET("/expenses", getExpensesHandler(&expenses))

	router.GET("/expenses/:id", getExpenseByIDHandler(&expenses))

	router.DELETE("/expenses/:id", deleteExpenseHandler(&expenses))

	router.PUT("/expenses/:id", updateExpenseHandler(&expenses))

	router.Run(":8080")
}
