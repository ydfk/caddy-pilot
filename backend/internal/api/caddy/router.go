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
	register(api, huma.Operation{OperationID: "get-caddy-settings", Method: http.MethodGet, Path: "/api/caddy/settings", Summary: "获取 Caddy 更新源设置"}, Settings)
	register(api, huma.Operation{OperationID: "update-caddy-settings", Method: http.MethodPut, Path: "/api/caddy/settings", Summary: "保存 Caddy 更新源设置"}, SaveSettings)
	register(api, huma.Operation{OperationID: "update-managed-caddy", Method: http.MethodPost, Path: "/api/caddy/update", Summary: "由后端下载并更新托管 Caddy"}, Update)
	register(api, huma.Operation{OperationID: "upload-managed-caddy", Method: http.MethodPost, Path: "/api/caddy/upload", Summary: "上传并更新托管 Caddy"}, Upload)
	register(api, huma.Operation{OperationID: "get-caddy-update-task", Method: http.MethodGet, Path: "/api/caddy/update-tasks/current", Summary: "获取当前 Caddy 更新任务"}, CurrentUpdateTask)
	register(api, huma.Operation{OperationID: "get-caddy-change-status", Method: http.MethodGet, Path: "/api/caddy/change-status", Summary: "检查是否存在未发布变更"}, ChangeStatus)
	register(api, huma.Operation{OperationID: "preview-caddy-config", Method: http.MethodPost, Path: "/api/caddy/preview", Summary: "预览完整 Caddy JSON"}, Preview)
	register(api, huma.Operation{OperationID: "preview-caddyfile", Method: http.MethodPost, Path: "/api/caddy/preview-caddyfile", Summary: "生成并校验只读 Caddyfile"}, PreviewCaddyfile)
	register(api, huma.Operation{OperationID: "validate-caddy-config", Method: http.MethodPost, Path: "/api/caddy/validate", Summary: "校验 Caddy JSON"}, Validate)
	register(api, huma.Operation{OperationID: "publish-caddy-config", Method: http.MethodPost, Path: "/api/caddy/publish", Summary: "发布 Caddy JSON"}, Publish)
	register(api, huma.Operation{OperationID: "get-current-caddy-config", Method: http.MethodGet, Path: "/api/caddy/current-config", Summary: "获取当前 Caddy JSON"}, CurrentConfig)
}

func register[I, O any](api huma.API, operation huma.Operation, handler func(context.Context, *I) (*O, error)) {
	operation.Tags = []string{"Caddy"}
	operation.Security = []map[string][]string{{auth.BearerAuthScheme: {}}}
	operation.Errors = []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusConflict, http.StatusBadGateway, http.StatusInternalServerError}
	huma.Register(api, operation, handler)
}
