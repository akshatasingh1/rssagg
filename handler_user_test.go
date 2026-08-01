package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/akshatasingh1/rssagg/internal/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// mockQuerier embeds database.Querier so it satisfies the full interface;
// only the methods exercised by a given test need to be overridden.
type mockQuerier struct {
	database.Querier
	createUserFunc     func(ctx context.Context, arg database.CreateUserParams) (database.User, error)
	getUserByEmailFunc func(ctx context.Context, email sql.NullString) (database.User, error)
}

func (m *mockQuerier) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	return m.createUserFunc(ctx, arg)
}

func (m *mockQuerier) GetUserByEmail(ctx context.Context, email sql.NullString) (database.User, error) {
	return m.getUserByEmailFunc(ctx, email)
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(hashed)
}

func doRequest(t *testing.T, handler http.HandlerFunc, body map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func TestHandlerCreateUser(t *testing.T) {
	tests := []struct {
		name       string
		body       map[string]string
		createUser func(ctx context.Context, arg database.CreateUserParams) (database.User, error)
		wantStatus int
	}{
		{
			name: "valid signup succeeds",
			body: map[string]string{"name": "Test User", "email": "test@example.com", "password": "password123"},
			createUser: func(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
				return database.User{
					ID:        uuid.New(),
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
					Name:      arg.Name,
					Email:     arg.Email,
					ApiKey:    "generated-api-key",
				}, nil
			},
			wantStatus: 201,
		},
		{
			name:       "missing password rejected before hitting DB",
			body:       map[string]string{"name": "Test User", "email": "test@example.com"},
			wantStatus: 400,
		},
		{
			name:       "short password rejected before hitting DB",
			body:       map[string]string{"name": "Test User", "email": "test@example.com", "password": "short"},
			wantStatus: 400,
		},
		{
			name:       "missing email rejected before hitting DB",
			body:       map[string]string{"name": "Test User", "password": "password123"},
			wantStatus: 400,
		},
		{
			name: "duplicate email rejected",
			body: map[string]string{"name": "Test User", "email": "taken@example.com", "password": "password123"},
			createUser: func(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
				return database.User{}, &pqDuplicateKeyError{}
			},
			wantStatus: 409,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiCfg := &apiConfig{DB: &mockQuerier{createUserFunc: tt.createUser}}
			w := doRequest(t, apiCfg.handlerCreateUser, tt.body)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantStatus == 201 {
				var got map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if got["api_key"] == nil || got["api_key"] == "" {
					t.Errorf("expected non-empty api_key in response, got %v", got["api_key"])
				}
				if _, exposed := got["password_hash"]; exposed {
					t.Errorf("password_hash must never be exposed in the API response")
				}
			}
		})
	}
}

func TestHandlerLogin(t *testing.T) {
	correctHash := hashPassword(t, "password123")

	tests := []struct {
		name          string
		body          map[string]string
		getUserByEmail func(ctx context.Context, email sql.NullString) (database.User, error)
		wantStatus    int
	}{
		{
			name: "correct password succeeds",
			body: map[string]string{"email": "test@example.com", "password": "password123"},
			getUserByEmail: func(ctx context.Context, email sql.NullString) (database.User, error) {
				return database.User{
					ID:           uuid.New(),
					Email:        sql.NullString{String: "test@example.com", Valid: true},
					PasswordHash: sql.NullString{String: correctHash, Valid: true},
					ApiKey:       "generated-api-key",
				}, nil
			},
			wantStatus: 200,
		},
		{
			name: "wrong password rejected",
			body: map[string]string{"email": "test@example.com", "password": "wrongpassword"},
			getUserByEmail: func(ctx context.Context, email sql.NullString) (database.User, error) {
				return database.User{
					Email:        sql.NullString{String: "test@example.com", Valid: true},
					PasswordHash: sql.NullString{String: correctHash, Valid: true},
				}, nil
			},
			wantStatus: 401,
		},
		{
			name: "unknown email rejected",
			body: map[string]string{"email": "nobody@example.com", "password": "password123"},
			getUserByEmail: func(ctx context.Context, email sql.NullString) (database.User, error) {
				return database.User{}, sql.ErrNoRows
			},
			wantStatus: 401,
		},
		{
			name: "account with no password set is rejected, not crashed",
			body: map[string]string{"email": "legacy@example.com", "password": "password123"},
			getUserByEmail: func(ctx context.Context, email sql.NullString) (database.User, error) {
				return database.User{
					Email:        sql.NullString{String: "legacy@example.com", Valid: true},
					PasswordHash: sql.NullString{Valid: false},
				}, nil
			},
			wantStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiCfg := &apiConfig{DB: &mockQuerier{getUserByEmailFunc: tt.getUserByEmail}}
			w := doRequest(t, apiCfg.handlerLogin, tt.body)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// pqDuplicateKeyError mimics the substring lib/pq puts in unique-violation errors,
// which handler_user.go checks for via strings.Contains(err.Error(), "duplicate key").
type pqDuplicateKeyError struct{}

func (e *pqDuplicateKeyError) Error() string {
	return `pq: duplicate key value violates unique constraint "users_email_key"`
}
