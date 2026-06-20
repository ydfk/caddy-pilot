package proxysite

import (
	"context"
	"net/http"

	"go-fiber-starter/internal/api/auth"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(api huma.API) {
	register(api, huma.Operation{OperationID: "list-proxy-sites", Method: http.MethodGet, Path: "/api/proxy-sites", Summary: "代理站点列表"}, List)
	register(api, huma.Operation{OperationID: "create-proxy-site", Method: http.MethodPost, Path: "/api/proxy-sites", Summary: "新增代理站点", DefaultStatus: http.StatusCreated}, Create)
	register(api, huma.Operation{OperationID: "get-proxy-site", Method: http.MethodGet, Path: "/api/proxy-sites/{id}", Summary: "代理站点详情"}, Get)
	register(api, huma.Operation{OperationID: "update-proxy-site", Method: http.MethodPut, Path: "/api/proxy-sites/{id}", Summary: "编辑代理站点"}, Update)
	register(api, huma.Operation{OperationID: "delete-proxy-site", Method: http.MethodDelete, Path: "/api/proxy-sites/{id}", Summary: "删除代理站点", DefaultStatus: http.StatusNoContent}, Delete)
	register(api, huma.Operation{OperationID: "clone-proxy-site", Method: http.MethodPost, Path: "/api/proxy-sites/{id}/clone", Summary: "克隆代理站点", DefaultStatus: http.StatusCreated}, Clone)
	register(api, huma.Operation{OperationID: "enable-proxy-site", Method: http.MethodPost, Path: "/api/proxy-sites/{id}/enable", Summary: "启用代理站点"}, Enable)
	register(api, huma.Operation{OperationID: "disable-proxy-site", Method: http.MethodPost, Path: "/api/proxy-sites/{id}/disable", Summary: "停用代理站点"}, Disable)
	register(api, huma.Operation{OperationID: "preview-proxy-site", Method: http.MethodPost, Path: "/api/proxy-sites/{id}/preview", Summary: "预览站点配置"}, Preview)
}

func register[I, O any](api huma.API, operation huma.Operation, handler func(context.Context, *I) (*O, error)) {
	operation.Tags = []string{"代理站点"}
	operation.Security = []map[string][]string{{auth.BearerAuthScheme: {}}}
	operation.Errors = []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError}
	huma.Register(api, operation, handler)
}
