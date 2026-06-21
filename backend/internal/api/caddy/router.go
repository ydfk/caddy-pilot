package caddy

import (
	"context"
	"net/http"

	"go-fiber-starter/internal/api/auth"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(api huma.API) {
	register(api, huma.Operation{OperationID: "get-caddy-status", Method: http.MethodGet, Path: "/api/caddy/status", Summary: "获取 Caddy 状态"}, Status)
	register(api, huma.Operation{OperationID: "get-caddy-version", Method: http.MethodGet, Path: "/api/caddy/version", Summary: "获取 Caddy 版本与更新信息"}, Version)
	register(api, huma.Operation{OperationID: "preview-caddy-config", Method: http.MethodPost, Path: "/api/caddy/preview", Summary: "预览完整 Caddy JSON"}, Preview)
	register(api, huma.Operation{OperationID: "validate-caddy-config", Method: http.MethodPost, Path: "/api/caddy/validate", Summary: "校验 Caddy JSON"}, Validate)
	register(api, huma.Operation{OperationID: "publish-caddy-config", Method: http.MethodPost, Path: "/api/caddy/publish", Summary: "发布 Caddy JSON"}, Publish)
	register(api, huma.Operation{OperationID: "get-current-caddy-config", Method: http.MethodGet, Path: "/api/caddy/current-config", Summary: "获取当前 Caddy JSON"}, CurrentConfig)
}

func register[I, O any](api huma.API, operation huma.Operation, handler func(context.Context, *I) (*O, error)) {
	operation.Tags = []string{"Caddy"}
	operation.Security = []map[string][]string{{auth.BearerAuthScheme: {}}}
	operation.Errors = []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusBadGateway, http.StatusInternalServerError}
	huma.Register(api, operation, handler)
}
