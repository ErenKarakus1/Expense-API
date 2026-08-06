package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func main() {
	router := gin.Default()
	router.GET("/health", health)
	router.Run(":8080")

}
