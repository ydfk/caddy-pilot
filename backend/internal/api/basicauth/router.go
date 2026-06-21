package basicauth

import (
	"context"
	"net/http"

	"go-fiber-starter/internal/api/auth"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(api huma.API) {
	register(api, huma.Operation{OperationID: "list-basic-auth-credentials", Method: http.MethodGet, Path: "/api/basic-auth-credentials", Summary: "密码本列表"}, List)
	register(api, huma.Operation{OperationID: "create-basic-auth-credential", Method: http.MethodPost, Path: "/api/basic-auth-credentials", Summary: "新增密码条目", DefaultStatus: http.StatusCreated}, Create)
	register(api, huma.Operation{OperationID: "update-basic-auth-credential", Method: http.MethodPut, Path: "/api/basic-auth-credentials/{id}", Summary: "编辑密码条目"}, Update)
	register(api, huma.Operation{OperationID: "delete-basic-auth-credential", Method: http.MethodDelete, Path: "/api/basic-auth-credentials/{id}", Summary: "删除密码条目", DefaultStatus: http.StatusNoContent}, Delete)
}

func register[I, O any](api huma.API, operation huma.Operation, handler func(context.Context, *I) (*O, error)) {
	operation.Tags = []string{"Basic Auth 密码本"}
	operation.Security = []map[string][]string{{auth.BearerAuthScheme: {}}}
	operation.Errors = []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError}
	huma.Register(api, operation, handler)
}
