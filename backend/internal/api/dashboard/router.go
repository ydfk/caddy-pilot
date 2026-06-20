package dashboard

import (
	"net/http"

	"go-fiber-starter/internal/api/auth"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-dashboard-summary", Method: http.MethodGet,
		Path: "/api/dashboard/summary", Summary: "获取仪表盘汇总", Tags: []string{"仪表盘"},
		Security: []map[string][]string{{auth.BearerAuthScheme: {}}},
		Errors:   []int{http.StatusUnauthorized, http.StatusInternalServerError},
	}, Summary)
}
