package logs

import (
	"net/http"

	"go-fiber-starter/internal/api/auth"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(api huma.API) {
	operation := huma.Operation{
		OperationID: "list-runtime-logs", Method: http.MethodGet, Path: "/api/logs",
		Summary: "读取系统或 Caddy 日志", Tags: []string{"Logs"},
		Security: []map[string][]string{{auth.BearerAuthScheme: {}}},
	}
	huma.Register(api, operation, List)
}
