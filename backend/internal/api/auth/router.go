/*
 * @Description: Copyright (c) ydfk. All rights reserved
 * @Author: ydfk
 * @Date: 2025-06-09 17:47:24
 * @LastEditors: ydfk
 * @LastEditTime: 2025-06-09 17:47:38
 */
package auth

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(api huma.API) {
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
