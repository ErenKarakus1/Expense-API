package validation

import (
	"errors"
	"net/mail"
	"strings"

	"github.com/ErenKarakus1/Expense-API/models"
)

func ValidateRegisterRequest(req models.RegisterRequest) error {
	name := strings.TrimSpace(req.Name)
	if len(name) < 3 {
		return errors.New("name must be at least 3 characters")
	}
	if len(name) > 50 {
		return errors.New("name can be at most 50 characters")
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(req.Email)); err != nil {
		return errors.New("invalid email")
	}
	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

func ValidateLoginRequest(req models.LoginRequest) error {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if len(email) == 0 {
		return errors.New("please provide an email")
	}
	if len(req.Password) == 0 {
		return errors.New("please provide a password")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("invalid email")
	}
	return nil
}
