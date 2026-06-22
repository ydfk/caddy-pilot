package systeminfo

import (
	"context"
	"net/http"

	"go-fiber-starter/internal/api/auth"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(api huma.API) {
	operation := huma.Operation{
		OperationID: "get-system-info",
		Method:      http.MethodGet,
		Path:        "/api/system/info",
		Summary:     "获取 CaddyPilot 系统信息",
		Tags:        []string{"System"},
		Security:    []map[string][]string{{auth.BearerAuthScheme: {}}},
	}
	huma.Register(api, operation, func(ctx context.Context, input *struct{}) (*InfoOutput, error) {
		return Info(ctx, input)
	})
}
