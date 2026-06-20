package auth

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestRegisterHashError(t *testing.T) {
	original := generateFromPassword
	t.Cleanup(func() {
		generateFromPassword = original
	})
	generateFromPassword = func(_ []byte, _ int) ([]byte, error) {
		return nil, errors.New("boom")
	}

	app := setupTestApp(t)
	resp := doJSONRequest(t, app, http.MethodPost, "/api/auth/register", fiber.Map{
		"username": "alice",
		"password": "secret",
	}, nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}
