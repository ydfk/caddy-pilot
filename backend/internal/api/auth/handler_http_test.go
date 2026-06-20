package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	model "go-fiber-starter/internal/model/user"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"
)

func setupTestApp(t *testing.T) *fiber.App {
	t.Helper()

	prevConfig := config.Current
	config.Current.Jwt.Secret = "test-secret"
	config.Current.Jwt.Expiration = 3600
	config.Current.App.Env = "test"
	config.Current.App.Port = "0"
	config.Current.Database.Path = ""
	config.IsProduction = false

	prevDB := db.DB
	dsn := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := gormDB.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	closeSQLDB(t, gormDB)
	db.DB = gormDB

	t.Cleanup(func() {
		config.Current = prevConfig
		db.DB = prevDB
	})

	app := fiber.New()
	humaConfig := huma.DefaultConfig("Test API", "1.0.0")
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		BearerAuthScheme: {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}
	api := humafiber.New(app, humaConfig)
	api.UseMiddleware(NewAuthMiddleware(api))
	RegisterRoutes(api)

	return app
}

func closeSQLDB(t *testing.T, gormDB *gorm.DB) {
	t.Helper()
	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
}

func doJSONRequest(t *testing.T, app *fiber.App, method, path string, body any, headers map[string]string) *http.Response {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	return resp
}

func decodeJSON[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer resp.Body.Close()
	var payload T
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return payload
}

func TestAuthFlowRegisterLoginProfile(t *testing.T) {
	app := setupTestApp(t)

	setupResp := doJSONRequest(t, app, http.MethodGet, "/api/auth/setup-status", nil, nil)
	if setupResp.StatusCode != http.StatusOK {
		t.Fatalf("setup status: %d", setupResp.StatusCode)
	}
	if decodeJSON[SetupStatusResponse](t, setupResp).Initialized {
		t.Fatal("fresh database should not be initialized")
	}

	registerResp := doJSONRequest(t, app, http.MethodPost, "/api/auth/register", fiber.Map{
		"username": "alice",
		"password": "pass123",
	}, nil)
	if registerResp.StatusCode != http.StatusCreated {
		t.Fatalf("register status: %d", registerResp.StatusCode)
	}
	registeredUser := decodeJSON[UserResponse](t, registerResp)
	if registeredUser.Username != "alice" {
		t.Fatalf("unexpected username: %s", registeredUser.Username)
	}
	initializedResp := doJSONRequest(t, app, http.MethodGet, "/api/auth/setup-status", nil, nil)
	if !decodeJSON[SetupStatusResponse](t, initializedResp).Initialized {
		t.Fatal("registered database should be initialized")
	}
	secondRegisterResp := doJSONRequest(t, app, http.MethodPost, "/api/auth/register", fiber.Map{
		"username": "bob",
		"password": "pass123",
	}, nil)
	defer secondRegisterResp.Body.Close()
	if secondRegisterResp.StatusCode != http.StatusConflict {
		t.Fatalf("second register status: %d", secondRegisterResp.StatusCode)
	}

	loginResp := doJSONRequest(t, app, http.MethodPost, "/api/auth/login", fiber.Map{
		"username": "alice",
		"password": "pass123",
	}, nil)
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login status: %d", loginResp.StatusCode)
	}
	token := decodeJSON[TokenResponse](t, loginResp)
	if token.Token == "" {
		t.Fatal("empty token")
	}

	profileResp := doJSONRequest(t, app, http.MethodGet, "/api/auth/profile", nil, map[string]string{
		"Authorization": "Bearer " + token.Token,
	})
	if profileResp.StatusCode != http.StatusOK {
		t.Fatalf("profile status: %d", profileResp.StatusCode)
	}
	profileUser := decodeJSON[UserResponse](t, profileResp)
	if profileUser.Username != "alice" {
		t.Fatalf("unexpected profile username: %s", profileUser.Username)
	}
}

func TestProfileRequiresAuth(t *testing.T) {
	app := setupTestApp(t)

	resp := doJSONRequest(t, app, http.MethodGet, "/api/auth/profile", nil, nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", resp.StatusCode)
	}
}

func TestOpenAPI31AndDocs(t *testing.T) {
	app := setupTestApp(t)

	specResp := doJSONRequest(t, app, http.MethodGet, "/openapi.yaml", nil, nil)
	defer specResp.Body.Close()
	if specResp.StatusCode != http.StatusOK {
		t.Fatalf("openapi status: %d", specResp.StatusCode)
	}
	spec, err := io.ReadAll(specResp.Body)
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}
	for _, expected := range []string{
		"openapi: 3.1.0",
		"/api/auth/setup-status:",
		"/api/auth/register:",
		"/api/auth/login:",
		"/api/auth/profile:",
		"bearerAuth:",
	} {
		if !strings.Contains(string(spec), expected) {
			t.Fatalf("openapi missing %q", expected)
		}
	}

	docsResp := doJSONRequest(t, app, http.MethodGet, "/docs", nil, nil)
	defer docsResp.Body.Close()
	if docsResp.StatusCode != http.StatusOK {
		t.Fatalf("docs status: %d", docsResp.StatusCode)
	}
}
