package auth

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-setup-status",
		Method:      http.MethodGet,
		Path:        "/api/auth/setup-status",
		Summary:     "获取初始化状态",
		Tags:        []string{"认证"},
		Errors:      []int{http.StatusInternalServerError},
	}, SetupStatus)

	huma.Register(api, huma.Operation{
		OperationID:   "register-user",
		Method:        http.MethodPost,
		Path:          "/api/auth/register",
		Summary:       "注册用户",
		Tags:          []string{"认证"},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusConflict, http.StatusInternalServerError},
	}, Register)

	huma.Register(api, huma.Operation{
		OperationID: "login-user",
		Method:      http.MethodPost,
		Path:        "/api/auth/login",
		Summary:     "用户登录",
		Tags:        []string{"认证"},
		Errors:      []int{http.StatusUnauthorized, http.StatusInternalServerError},
	}, Login)

	huma.Register(api, huma.Operation{
		OperationID: "get-passkey-status",
		Method:      http.MethodGet,
		Path:        "/api/auth/passkeys/status",
		Summary:     "获取 Passkey 登录状态",
		Tags:        []string{"认证"},
	}, PasskeyStatus)

	huma.Register(api, huma.Operation{
		OperationID: "begin-passkey-login",
		Method:      http.MethodPost,
		Path:        "/api/auth/passkeys/login/options",
		Summary:     "创建 Passkey 登录挑战",
		Tags:        []string{"认证"},
		Errors:      []int{http.StatusBadRequest, http.StatusServiceUnavailable},
	}, BeginPasskeyLogin)

	huma.Register(api, huma.Operation{
		OperationID: "finish-passkey-login",
		Method:      http.MethodPost,
		Path:        "/api/auth/passkeys/login/verify",
		Summary:     "验证 Passkey 并登录",
		Tags:        []string{"认证"},
		Errors:      []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusServiceUnavailable},
	}, FinishPasskeyLogin)

	huma.Register(api, huma.Operation{
		OperationID: "list-passkeys",
		Method:      http.MethodGet,
		Path:        "/api/auth/passkeys",
		Summary:     "列出当前用户的 Passkey",
		Tags:        []string{"认证"},
		Security:    []map[string][]string{{BearerAuthScheme: {}}},
	}, ListPasskeys)

	huma.Register(api, huma.Operation{
		OperationID: "begin-passkey-registration",
		Method:      http.MethodPost,
		Path:        "/api/auth/passkeys/register/options",
		Summary:     "创建 Passkey 注册挑战",
		Tags:        []string{"认证"},
		Security:    []map[string][]string{{BearerAuthScheme: {}}},
		Errors:      []int{http.StatusBadRequest, http.StatusServiceUnavailable},
	}, BeginPasskeyRegistration)

	huma.Register(api, huma.Operation{
		OperationID: "finish-passkey-registration",
		Method:      http.MethodPost,
		Path:        "/api/auth/passkeys/register/verify",
		Summary:     "验证并保存 Passkey",
		Tags:        []string{"认证"},
		Security:    []map[string][]string{{BearerAuthScheme: {}}},
		Errors:      []int{http.StatusBadRequest, http.StatusServiceUnavailable},
	}, FinishPasskeyRegistration)

	huma.Register(api, huma.Operation{
		OperationID: "rename-passkey",
		Method:      http.MethodPatch,
		Path:        "/api/auth/passkeys/{id}",
		Summary:     "重命名 Passkey",
		Tags:        []string{"认证"},
		Security:    []map[string][]string{{BearerAuthScheme: {}}},
	}, RenamePasskey)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-passkey",
		Method:        http.MethodDelete,
		Path:          "/api/auth/passkeys/{id}",
		Summary:       "删除 Passkey",
		Tags:          []string{"认证"},
		Security:      []map[string][]string{{BearerAuthScheme: {}}},
		DefaultStatus: http.StatusNoContent,
	}, DeletePasskey)

	huma.Register(api, huma.Operation{
		OperationID: "get-user-profile",
		Method:      http.MethodGet,
		Path:        "/api/auth/profile",
		Summary:     "获取当前用户",
		Tags:        []string{"认证"},
		Security: []map[string][]string{
			{BearerAuthScheme: {}},
		},
		Errors: []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError},
	}, Profile)
}
