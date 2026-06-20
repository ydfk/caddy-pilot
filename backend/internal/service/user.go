package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	model "go-fiber-starter/internal/model/user"
	"go-fiber-starter/pkg/config"
)

func GenerateJWT(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   user.Id,
		"user_name": user.Username,
		"exp":       time.Now().Add(time.Duration(config.Current.Jwt.Expiration) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Current.Jwt.Secret))
}

func AuthenticateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("不支持的 JWT 签名算法")
		}
		return []byte(config.Current.Jwt.Secret), nil
	}, jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return "", errors.New("JWT 无效或已过期")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("JWT 声明无效")
	}

	return parseUserIDClaim(claims)
}

func parseUserIDClaim(claims jwt.MapClaims) (string, error) {
	value, ok := claims["user_id"]
	if !ok || value == nil {
		return "", errors.New("user_id claim missing")
	}

	switch typed := value.(type) {
	case string:
		if typed == "" {
			return "", errors.New("user_id claim missing")
		}
		return typed, nil
	case uuid.UUID:
		return typed.String(), nil
	case []byte:
		if len(typed) == 0 {
			return "", errors.New("user_id claim missing")
		}
		return string(typed), nil
	default:
		return "", errors.New("user_id claim invalid")
	}
}
