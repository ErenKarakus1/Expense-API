package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const baseURL = "http://localhost:8080"

const testPassword = "password123"

type Expense struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	AmountCents int64  `json:"amount_cents"`
	Category    string `json:"category"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreatedUser struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func TestHealth(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/health", "", nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)
}

func TestRegisterSuccess(t *testing.T) {
	email := uniqueEmail("register-success")

	resp := register(t, email, testPassword)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusCreated)

	var user CreatedUser
	decodeJSON(t, resp, &user)

	if user.ID == "" {
		t.Fatal("id should not be empty")
	}
	if user.Email != email {
		t.Fatalf("expected email %q got %q", email, user.Email)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	email := uniqueEmail("duplicate")

	resp := register(t, email, testPassword)
	resp.Body.Close()
	assertStatus(t, resp, http.StatusCreated)

	resp = register(t, email, testPassword)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusConflict)
}

func TestRegisterValidation(t *testing.T) {
	testCases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "invalid email",
			body: map[string]any{
				"name":     "Eren",
				"email":    "not-an-email",
				"password": testPassword,
			},
		},
		{
			name: "short password",
			body: map[string]any{
				"name":     "Eren",
				"email":    uniqueEmail("short-password"),
				"password": "123",
			},
		},
		{
			name: "short name",
			body: map[string]any{
				"name":     "Er",
				"email":    uniqueEmail("short-name"),
				"password": testPassword,
			},
		},
		{
			name: "long name",
			body: map[string]any{
				"name":     strings.Repeat("a", 51),
				"email":    uniqueEmail("long-name"),
				"password": testPassword,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := doRequest(t, http.MethodPost, "/register", "", tc.body)
			defer resp.Body.Close()

			assertStatus(t, resp, http.StatusBadRequest)
		})
	}
}

func TestRegisterInvalidJSON(t *testing.T) {
	resp := doRawRequest(t, http.MethodPost, "/register", "", []byte(`{"name":`))
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusBadRequest)
}

func TestLoginSuccess(t *testing.T) {
	email := uniqueEmail("login-success")
	registerUser(t, email, testPassword)

	resp := login(t, email, testPassword)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)

	var loginResponse LoginResponse
	decodeJSON(t, resp, &loginResponse)

	if loginResponse.Token == "" {
		t.Fatal("token should not be empty")
	}
}

func TestLoginFailures(t *testing.T) {
	email := uniqueEmail("login-failures")
	registerUser(t, email, testPassword)

	testCases := []struct {
		name string
		body map[string]any
		code int
	}{
		{
			name: "wrong password",
			body: map[string]any{
				"email":    email,
				"password": "wrongpassword",
			},
			code: http.StatusUnauthorized,
		},
		{
			name: "user not found",
			body: map[string]any{
				"email":    uniqueEmail("missing-user"),
				"password": testPassword,
			},
			code: http.StatusUnauthorized,
		},
		{
			name: "invalid email",
			body: map[string]any{
				"email":    "not-email",
				"password": testPassword,
			},
			code: http.StatusBadRequest,
		},
		{
			name: "missing password",
			body: map[string]any{
				"email": email,
			},
			code: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := doRequest(t, http.MethodPost, "/login", "", tc.body)
			defer resp.Body.Close()

			assertStatus(t, resp, tc.code)
		})
	}
}

func TestLoginInvalidJSON(t *testing.T) {
	resp := doRawRequest(t, http.MethodPost, "/login", "", []byte(`{"email":`))
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusBadRequest)
}

func TestExpensesRequireAuth(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/expenses", "", nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusUnauthorized)
}

func TestInvalidToken(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/expenses", "not-a-valid-token", nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusUnauthorized)
}

func TestInvalidAuthHeaderFormat(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, baseURL+"/expenses", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "not-a-bearer-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusUnauthorized)
}

func TestFullExpenseFlow(t *testing.T) {
	token := registerAndLogin(t, uniqueEmail("expense-flow"))

	created := createExpense(t, token, 1500, "Food", "Pizza")
	if created.ID == "" {
		t.Fatal("id should not be empty")
	}
	if created.UserID == "" {
		t.Fatal("user_id should not be empty")
	}

	resp := doRequest(t, http.MethodGet, "/expenses/"+created.ID, token, nil)
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusOK)

	var found Expense
	decodeJSON(t, resp, &found)
	if found.ID != created.ID {
		t.Fatalf("expected expense id %q got %q", created.ID, found.ID)
	}

	resp = doRequest(t, http.MethodGet, "/expenses", token, nil)
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusOK)

	var expenses []Expense
	decodeJSON(t, resp, &expenses)
	if len(expenses) == 0 {
		t.Fatal("expected at least one expense")
	}

	updateBody := map[string]any{
		"amount_cents": 3000,
		"category":     "Updated",
		"description":  "Updated Expense",
	}
	resp = doRequest(t, http.MethodPut, "/expenses/"+created.ID, token, updateBody)
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusOK)

	var updated Expense
	decodeJSON(t, resp, &updated)
	if updated.AmountCents != 3000 {
		t.Fatalf("expected updated amount 3000 got %d", updated.AmountCents)
	}

	resp = doRequest(t, http.MethodDelete, "/expenses/"+created.ID, token, nil)
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusNoContent)

	resp = doRequest(t, http.MethodGet, "/expenses/"+created.ID, token, nil)
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusNotFound)
}

func TestExpenseValidation(t *testing.T) {
	token := registerAndLogin(t, uniqueEmail("expense-validation"))

	testCases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "zero amount",
			body: map[string]any{
				"amount_cents": 0,
				"category":     "Food",
			},
		},
		{
			name: "negative amount",
			body: map[string]any{
				"amount_cents": -100,
				"category":     "Food",
			},
		},
		{
			name: "category too long",
			body: map[string]any{
				"amount_cents": 100,
				"category":     strings.Repeat("a", 51),
			},
		},
		{
			name: "description too long",
			body: map[string]any{
				"amount_cents": 100,
				"description":  strings.Repeat("a", 501),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := doRequest(t, http.MethodPost, "/expenses", token, tc.body)
			defer resp.Body.Close()

			assertStatus(t, resp, http.StatusBadRequest)
		})
	}
}

func TestExpenseInvalidJSON(t *testing.T) {
	token := registerAndLogin(t, uniqueEmail("expense-invalid-json"))

	resp := doRawRequest(t, http.MethodPost, "/expenses", token, []byte(`{"amount_cents":`))
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusBadRequest)
}

func TestExpenseInvalidUUIDs(t *testing.T) {
	token := registerAndLogin(t, uniqueEmail("invalid-uuid"))

	endpoints := []struct {
		method string
		path   string
		body   any
	}{
		{method: http.MethodGet, path: "/expenses/not-a-uuid"},
		{method: http.MethodPut, path: "/expenses/not-a-uuid", body: validExpenseBody()},
		{method: http.MethodDelete, path: "/expenses/not-a-uuid"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.method, func(t *testing.T) {
			resp := doRequest(t, endpoint.method, endpoint.path, token, endpoint.body)
			defer resp.Body.Close()

			assertStatus(t, resp, http.StatusBadRequest)
		})
	}
}

func TestExpenseNotFound(t *testing.T) {
	token := registerAndLogin(t, uniqueEmail("not-found"))
	path := "/expenses/" + uuid.New().String()

	resp := doRequest(t, http.MethodGet, path, token, nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusNotFound)
}

func TestGetExpensesReturnsArray(t *testing.T) {
	token := registerAndLogin(t, uniqueEmail("returns-array"))

	resp := doRequest(t, http.MethodGet, "/expenses", token, nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)

	var expenses []Expense
	decodeJSON(t, resp, &expenses)

	if expenses == nil {
		t.Fatal("expected [] not null")
	}
}

func TestExpenseOrdering(t *testing.T) {
	token := registerAndLogin(t, uniqueEmail("ordering"))

	createExpense(t, token, 100, "Test", "Ordering")
	time.Sleep(10 * time.Millisecond)
	createExpense(t, token, 200, "Test", "Ordering")
	time.Sleep(10 * time.Millisecond)
	createExpense(t, token, 300, "Test", "Ordering")

	resp := doRequest(t, http.MethodGet, "/expenses", token, nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)

	var expenses []Expense
	decodeJSON(t, resp, &expenses)

	if len(expenses) < 3 {
		t.Fatal("not enough expenses returned")
	}
	if expenses[0].AmountCents != 300 {
		t.Fatalf("expected newest expense first, got amount %d", expenses[0].AmountCents)
	}
}

func TestExpenseUserIsolation(t *testing.T) {
	firstToken := registerAndLogin(t, uniqueEmail("owner"))
	secondToken := registerAndLogin(t, uniqueEmail("other"))

	created := createExpense(t, firstToken, 2500, "Travel", "Train")
	path := "/expenses/" + created.ID

	resp := doRequest(t, http.MethodGet, path, secondToken, nil)
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusNotFound)

	resp = doRequest(t, http.MethodPut, path, secondToken, validExpenseBody())
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusNotFound)

	resp = doRequest(t, http.MethodDelete, path, secondToken, nil)
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusNotFound)

	resp = doRequest(t, http.MethodGet, path, firstToken, nil)
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusOK)
}

func createExpense(t *testing.T, token string, amount int64, category string, description string) Expense {
	t.Helper()

	body := map[string]any{
		"amount_cents": amount,
		"category":     category,
		"description":  description,
	}

	resp := doRequest(t, http.MethodPost, "/expenses", token, body)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusCreated)

	var expense Expense
	decodeJSON(t, resp, &expense)

	return expense
}

func validExpenseBody() map[string]any {
	return map[string]any{
		"amount_cents": 1000,
		"category":     "Test",
		"description":  "Valid expense",
	}
}

func registerAndLogin(t *testing.T, email string) string {
	t.Helper()

	registerUser(t, email, testPassword)

	resp := login(t, email, testPassword)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)

	var loginResponse LoginResponse
	decodeJSON(t, resp, &loginResponse)

	if loginResponse.Token == "" {
		t.Fatal("token should not be empty")
	}
	return loginResponse.Token
}

func registerUser(t *testing.T, email string, password string) {
	t.Helper()

	resp := register(t, email, password)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusCreated)
}

func register(t *testing.T, email string, password string) *http.Response {
	t.Helper()

	body := map[string]any{
		"name":     "Test User",
		"email":    email,
		"password": password,
	}

	return doRequest(t, http.MethodPost, "/register", "", body)
}

func login(t *testing.T, email string, password string) *http.Response {
	t.Helper()

	body := map[string]any{
		"email":    email,
		"password": password,
	}

	return doRequest(t, http.MethodPost, "/login", "", body)
}

func doRequest(t *testing.T, method string, path string, token string, body any) *http.Response {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}

	return doRawRequest(t, method, path, token, payload)
}

func doRawRequest(t *testing.T, method string, path string, token string, payload []byte) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, baseURL+path, bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	return resp
}

func decodeJSON(t *testing.T, resp *http.Response, target any) {
	t.Helper()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatal(err)
	}
}

func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()

	if resp.StatusCode != expected {
		t.Fatalf("expected %d got %d", expected, resp.StatusCode)
	}
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf(
		"%s-%d-%s@example.com",
		prefix,
		time.Now().UnixNano(),
		uuid.NewString(),
	)
}
