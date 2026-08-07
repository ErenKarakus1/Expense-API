package handlers

import (
	"errors"
	"net/http"

	"github.com/ErenKarakus1/Expense-API/auth"
	"github.com/ErenKarakus1/Expense-API/config"
	"github.com/ErenKarakus1/Expense-API/models"
	"github.com/ErenKarakus1/Expense-API/repository"
	"github.com/ErenKarakus1/Expense-API/validation"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := buildUser(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		createdUser, err := repository.CreateUser(c.Request.Context(), pool, user)
		if err != nil {
			if errors.Is(err, repository.ErrEmailExists) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, createdUser)
	}
}

func LoginHandler(pool *pgxpool.Pool, cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginUser models.LoginRequest
		if err := c.ShouldBindJSON(&loginUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if err := validation.ValidateLoginRequest(loginUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user, err := repository.FindUserByEmail(c.Request.Context(), pool, loginUser.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		if err := auth.ComparePassword(user.PasswordHash, loginUser.Password); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		token, err := auth.GenerateToken(user.ID, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": repository.ErrInternalServer.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}
