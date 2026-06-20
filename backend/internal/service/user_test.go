package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"go-fiber-starter/internal/model/base"
	model "go-fiber-starter/internal/model/user"
	"go-fiber-starter/pkg/config"
)

func TestGenerateJWTUsesConfigExpiration(t *testing.T) {
	setTestJWTConfig(t)

	user := &model.User{
		BaseModel: base.BaseModel{Id: uuid.New()},
		Username:  "test-user",
	}

	start := time.Now()
	tokenString, err := GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT returned error: %v", err)
	}

	parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(config.Current.Jwt.Secret), nil
	})
	if err != nil {
		t.Fatalf("Parse token error: %v", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("Expected MapClaims, got %T", parsed.Claims)
	}
	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil {
		t.Fatalf("exp claim invalid: %v", err)
	}

	expectedMin := start.Unix() + int64(config.Current.Jwt.Expiration)
	expectedMax := time.Now().Unix() + int64(config.Current.Jwt.Expiration)
	if exp.Unix() < expectedMin || exp.Unix() > expectedMax {
		t.Fatalf("exp claim %d not within [%d, %d]", exp.Unix(), expectedMin, expectedMax)
	}
}

func TestAuthenticateJWT(t *testing.T) {
	setTestJWTConfig(t)
	userID := uuid.New()
	token := signedToken(t, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	parsedID, err := AuthenticateJWT(token)
	if err != nil {
		t.Fatalf("AuthenticateJWT returned error: %v", err)
	}
	if parsedID != userID.String() {
		t.Fatalf("AuthenticateJWT id %s != %s", parsedID, userID)
	}
}

func TestParseUserIDClaimUUID(t *testing.T) {
	userID := uuid.New()
	parsed, err := parseUserIDClaim(jwt.MapClaims{"user_id": userID})
	if err != nil {
		t.Fatalf("parseUserIDClaim returned error: %v", err)
	}
	if parsed != userID.String() {
		t.Fatalf("parseUserIDClaim id %s != %s", parsed, userID)
	}
}

func TestParseUserIDClaimBytes(t *testing.T) {
	userID := uuid.New()
	parsed, err := parseUserIDClaim(jwt.MapClaims{"user_id": []byte(userID.String())})
	if err != nil {
		t.Fatalf("parseUserIDClaim returned error: %v", err)
	}
	if parsed != userID.String() {
		t.Fatalf("parseUserIDClaim id %s != %s", parsed, userID)
	}
}

func TestAuthenticateJWTInvalidClaims(t *testing.T) {
	setTestJWTConfig(t)

	for name, claims := range map[string]jwt.MapClaims{
		"missing": {"exp": time.Now().Add(time.Hour).Unix()},
		"type":    {"user_id": 123, "exp": time.Now().Add(time.Hour).Unix()},
	} {
		t.Run(name, func(t *testing.T) {
			userID, err := AuthenticateJWT(signedToken(t, claims))
			if err == nil {
				t.Fatalf("AuthenticateJWT expected error, got user ID %s", userID)
			}
		})
	}
}

func TestAuthenticateJWTExpiredToken(t *testing.T) {
	setTestJWTConfig(t)
	token := signedToken(t, jwt.MapClaims{
		"user_id": uuid.NewString(),
		"exp":     time.Now().Add(-time.Minute).Unix(),
	})

	userID, err := AuthenticateJWT(token)
	if err == nil {
		t.Fatalf("AuthenticateJWT expected error, got user ID %s", userID)
	}
}

func signedToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Current.Jwt.Secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return tokenString
}

func setTestJWTConfig(t *testing.T) {
	t.Helper()

	prevSecret := config.Current.Jwt.Secret
	prevExpiration := config.Current.Jwt.Expiration
	config.Current.Jwt.Secret = "test-secret"
	config.Current.Jwt.Expiration = 3600

	t.Cleanup(func() {
		config.Current.Jwt.Secret = prevSecret
		config.Current.Jwt.Expiration = prevExpiration
	})
}
