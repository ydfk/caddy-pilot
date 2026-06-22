package certificate

import (
	"context"
	"net/http"

	"go-fiber-starter/internal/api/auth"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(api huma.API) {
	register(api, huma.Operation{OperationID: "list-certificate-profiles", Method: http.MethodGet, Path: "/api/certificates", Summary: "证书配置列表"}, List)
	register(api, huma.Operation{OperationID: "create-certificate-profile", Method: http.MethodPost, Path: "/api/certificates", Summary: "新增证书配置", DefaultStatus: http.StatusCreated}, Create)
	register(api, huma.Operation{OperationID: "update-certificate-profile", Method: http.MethodPut, Path: "/api/certificates/{id}", Summary: "编辑证书配置"}, Update)
	register(api, huma.Operation{OperationID: "delete-certificate-profile", Method: http.MethodDelete, Path: "/api/certificates/{id}", Summary: "删除证书配置", DefaultStatus: http.StatusNoContent}, Delete)
}

func register[I, O any](api huma.API, operation huma.Operation, handler func(context.Context, *I) (*O, error)) {
	operation.Tags = []string{"证书"}
	operation.Security = []map[string][]string{{auth.BearerAuthScheme: {}}}
	operation.Errors = []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError}
	huma.Register(api, operation, handler)
}
