package auth

import (
	"context"
	"net/http"
	"strings"

	"go-fiber-starter/internal/service"
	"go-fiber-starter/pkg/logger"

	"github.com/danielgtaylor/huma/v2"
)

const BearerAuthScheme = "bearerAuth"

type userIDContextKey struct{}

func NewAuthMiddleware(api huma.API) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if !requiresBearerAuth(ctx.Operation()) {
			next(ctx)
			return
		}

		token := bearerToken(ctx.Header("Authorization"))
		if token == "" {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "请提供 Bearer Token")
			return
		}

		userID, err := service.AuthenticateJWT(token)
		if err != nil {
			logger.Error("JWT 验证失败: %v", err)
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "认证失败，请重新登录")
			return
		}

		next(huma.WithValue(ctx, userIDContextKey{}, userID))
	}
}

func requiresBearerAuth(operation *huma.Operation) bool {
	for _, security := range operation.Security {
		if _, ok := security[BearerAuthScheme]; ok {
			return true
		}
	}
	return false
}

func bearerToken(header string) string {
	scheme, token, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

func userIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey{}).(string)
	return userID, ok && userID != ""
}
