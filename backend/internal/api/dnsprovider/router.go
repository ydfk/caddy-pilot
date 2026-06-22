package dnsprovider

import (
	"context"
	"net/http"

	"go-fiber-starter/internal/api/auth"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(api huma.API) {
	register(api, huma.Operation{OperationID: "list-dns-providers", Method: http.MethodGet, Path: "/api/dns-providers", Summary: "DNS Provider 列表"}, List)
	register(api, huma.Operation{OperationID: "create-dns-provider", Method: http.MethodPost, Path: "/api/dns-providers", Summary: "新增 DNS Provider", DefaultStatus: http.StatusCreated}, Create)
	register(api, huma.Operation{OperationID: "update-dns-provider", Method: http.MethodPut, Path: "/api/dns-providers/{id}", Summary: "编辑 DNS Provider"}, Update)
	register(api, huma.Operation{OperationID: "delete-dns-provider", Method: http.MethodDelete, Path: "/api/dns-providers/{id}", Summary: "删除 DNS Provider", DefaultStatus: http.StatusNoContent}, Delete)
}

func register[I, O any](api huma.API, operation huma.Operation, handler func(context.Context, *I) (*O, error)) {
	operation.Tags = []string{"DNS Provider"}
	operation.Security = []map[string][]string{{auth.BearerAuthScheme: {}}}
	operation.Errors = []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError}
	huma.Register(api, operation, handler)
}
