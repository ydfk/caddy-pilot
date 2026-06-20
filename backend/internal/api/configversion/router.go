package configversion

import (
	"context"
	"net/http"

	"go-fiber-starter/internal/api/auth"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(api huma.API) {
	register(api, huma.Operation{OperationID: "list-config-versions", Method: http.MethodGet, Path: "/api/config-versions", Summary: "配置版本列表"}, List)
	register(api, huma.Operation{OperationID: "get-config-version", Method: http.MethodGet, Path: "/api/config-versions/{id}", Summary: "配置版本详情"}, Get)
	register(api, huma.Operation{OperationID: "rollback-config-version", Method: http.MethodPost, Path: "/api/config-versions/{id}/rollback", Summary: "回滚配置版本"}, Rollback)
}

func register[I, O any](api huma.API, operation huma.Operation, handler func(context.Context, *I) (*O, error)) {
	operation.Tags = []string{"配置版本"}
	operation.Security = []map[string][]string{{auth.BearerAuthScheme: {}}}
	operation.Errors = []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusBadGateway, http.StatusInternalServerError}
	huma.Register(api, operation, handler)
}
