package proxysite

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-fiber-starter/internal/api/auth"
	"go-fiber-starter/internal/model/base"
	model "go-fiber-starter/internal/model/proxysite"
	userModel "go-fiber-starter/internal/model/user"
	"go-fiber-starter/internal/service"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCustomSiteSaveRunsCaddyValidation(t *testing.T) {
	original := validateCaddyConfig
	defer func() { validateCaddyConfig = original }()
	called := false
	validateCaddyConfig = func(_ context.Context, _ []byte) error {
		called = true
		return errors.New("模块配置无效")
	}

	site := model.ProxySite{
		ConfigMode: "custom", CustomFormat: "json",
		CustomConfig: `{"match":[{"host":["example.com"]}],"handle":[]}`,
		EnableHTTPS:  true,
	}
	err := validateCustomSite(context.Background(), site)
	if !called || err == nil || !strings.Contains(err.Error(), "模块配置无效") {
		t.Fatalf("自定义站点保存应执行 Caddy 校验: called=%v err=%v", called, err)
	}
}

func TestProxySiteLifecycle(t *testing.T) {
	app, token := setupProxySiteTestApp(t)

	unauthorized := proxySiteRequest(t, app, http.MethodGet, "/api/proxy-sites", nil, "")
	if unauthorized.StatusCode != http.StatusUnauthorized {
		t.Fatalf("未认证请求状态码为 %d", unauthorized.StatusCode)
	}
	_ = unauthorized.Body.Close()

	payload := validSitePayload("示例站点")
	createdResponse := proxySiteRequest(t, app, http.MethodPost, "/api/proxy-sites", payload, token)
	if createdResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createdResponse.Body)
		t.Fatalf("创建站点状态码为 %d: %s", createdResponse.StatusCode, body)
	}
	created := decodeProxySiteResponse(t, createdResponse)
	if created.Name != "示例站点" || !created.Enabled {
		t.Fatalf("创建结果不正确: %+v", created)
	}
	listResponse := proxySiteRequest(t, app, http.MethodGet, "/api/proxy-sites?page=1&page_size=1", nil, token)
	var page SiteListResponse
	decodeProxySiteJSON(t, listResponse, &page)
	if page.Total != 1 || page.PageSize != 1 || len(page.Items) != 1 {
		t.Fatalf("代理站点分页不正确: %+v", page)
	}

	payload["name"] = "更新站点"
	updatedResponse := proxySiteRequest(t, app, http.MethodPut, "/api/proxy-sites/"+created.ID.String(), payload, token)
	if updatedResponse.StatusCode != http.StatusOK {
		t.Fatalf("更新站点状态码为 %d", updatedResponse.StatusCode)
	}
	updated := decodeProxySiteResponse(t, updatedResponse)
	if updated.Name != "更新站点" {
		t.Fatalf("更新后的名称为 %s", updated.Name)
	}

	clonedResponse := proxySiteRequest(t, app, http.MethodPost, "/api/proxy-sites/"+created.ID.String()+"/clone", map[string]any{}, token)
	if clonedResponse.StatusCode != http.StatusCreated {
		t.Fatalf("克隆站点状态码为 %d", clonedResponse.StatusCode)
	}
	cloned := decodeProxySiteResponse(t, clonedResponse)
	if cloned.Enabled || cloned.ID == created.ID {
		t.Fatalf("克隆站点未默认停用或复用了原 ID: %+v", cloned)
	}

	emptyNameCloneResponse := proxySiteRequest(t, app, http.MethodPost, "/api/proxy-sites/"+created.ID.String()+"/clone", map[string]any{"name": ""}, token)
	if emptyNameCloneResponse.StatusCode != http.StatusCreated {
		t.Fatalf("空名称克隆站点状态码为 %d", emptyNameCloneResponse.StatusCode)
	}
	emptyNameClone := decodeProxySiteResponse(t, emptyNameCloneResponse)
	if emptyNameClone.Name != "更新站点 副本" {
		t.Fatalf("空名称克隆未沿用默认名称: %+v", emptyNameClone)
	}

	previewResponse := proxySiteRequest(t, app, http.MethodPost, "/api/proxy-sites/"+created.ID.String()+"/preview", nil, token)
	if previewResponse.StatusCode != http.StatusOK {
		t.Fatalf("预览站点状态码为 %d", previewResponse.StatusCode)
	}
	var preview PreviewResponse
	decodeProxySiteJSON(t, previewResponse, &preview)
	if !bytes.Contains(preview.CaddyJSON, []byte("reverse_proxy")) || !strings.Contains(preview.Caddyfile, "reverse_proxy") {
		t.Fatalf("预览配置缺少 JSON 或 Caddyfile reverse_proxy: %+v", preview)
	}

	deleteResponse := proxySiteRequest(t, app, http.MethodDelete, "/api/proxy-sites/"+created.ID.String(), nil, token)
	if deleteResponse.StatusCode != http.StatusNoContent {
		t.Fatalf("删除站点状态码为 %d", deleteResponse.StatusCode)
	}
	_ = deleteResponse.Body.Close()

	var deleted model.ProxySite
	if err := db.DB.Unscoped().First(&deleted, "id = ?", created.ID).Error; err != nil {
		t.Fatalf("读取软删除站点失败: %v", err)
	}
	if !deleted.DeletedAt.Valid {
		t.Fatal("站点未执行软删除")
	}
}

func TestImportNginxCreatesDisabledSites(t *testing.T) {
	app, token := setupProxySiteTestApp(t)
	payload := map[string]string{"config": `server {
		listen 443 ssl;
		server_name imported.example.com;
		location / { proxy_pass http://127.0.0.1:9000; }
	}`}
	response := proxySiteRequest(t, app, http.MethodPost, "/api/proxy-sites/import/nginx", payload, token)
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("导入 Nginx 状态码为 %d", response.StatusCode)
	}
	var imported NginxImportResponse
	decodeProxySiteJSON(t, response, &imported)
	if len(imported.Sites) != 1 || imported.Sites[0].Enabled || imported.Sites[0].Domains[0] != "imported.example.com" {
		t.Fatalf("导入结果不正确: %+v", imported)
	}
}

func TestCreateStaticSiteWithoutUpstream(t *testing.T) {
	app, token := setupProxySiteTestApp(t)
	payload := validSitePayload("静态站点")
	payload["site_type"] = "static"
	payload["upstreams"] = []string{}
	payload["root_path"] = "/var/www/example"
	payload["enable_security_headers"] = true
	payload["enable_asset_cache"] = true
	response := proxySiteRequest(t, app, http.MethodPost, "/api/proxy-sites", payload, token)
	if response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("创建静态站点状态码为 %d: %s", response.StatusCode, body)
	}
	created := decodeProxySiteResponse(t, response)
	if created.SiteType != "static" || created.RootPath != "/var/www/example" || len(created.Upstreams) != 0 {
		t.Fatalf("静态站点结果不正确: %+v", created)
	}
}

func setupProxySiteTestApp(t *testing.T) (*fiber.App, string) {
	t.Helper()
	previousDB := db.DB
	previousConfig := config.Current
	database, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := database.AutoMigrate(&model.ProxySite{}); err != nil {
		t.Fatalf("迁移测试数据库失败: %v", err)
	}
	db.DB = database
	config.Current.Jwt.Secret = "proxy-site-test-secret"
	config.Current.Jwt.Expiration = 3600
	t.Cleanup(func() {
		db.DB = previousDB
		config.Current = previousConfig
	})

	app := fiber.New()
	humaConfig := huma.DefaultConfig("代理站点测试", "1.0.0")
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		auth.BearerAuthScheme: {Type: "http", Scheme: "bearer", BearerFormat: "JWT"},
	}
	api := humafiber.New(app, humaConfig)
	api.UseMiddleware(auth.NewAuthMiddleware(api))
	RegisterRoutes(api)

	token, err := service.GenerateJWT(&userModel.User{BaseModel: base.BaseModel{Id: uuid.New()}})
	if err != nil {
		t.Fatalf("生成测试 Token 失败: %v", err)
	}
	return app, token
}

func validSitePayload(name string) map[string]any {
	return map[string]any{
		"name": name, "description": "测试", "domains": []string{"example.com"},
		"upstreams": []string{"127.0.0.1:3000"}, "enable_https": true,
		"upstream_type": "http", "upstream_tls_server_name": "",
		"upstream_tls_insecure_skip_verify": false,
		"force_https":                       true, "enable_gzip": true, "enable_log": false,
		"certificate_type": "single", "certificate_domain": "",
		"acme_challenge_type": "http", "dns_provider": "alidns",
		"enable_ws": true, "request_headers": map[string]string{},
		"response_headers": map[string]string{}, "basic_auth_enabled": false,
		"basic_auth_users": map[string]string{}, "allowed_ips": []string{},
		"basic_auth_credential_ids": []string{},
		"advanced_json":             "", "enabled": true,
	}
}

func proxySiteRequest(t *testing.T, app *fiber.App, method, path string, body any, token string) *http.Response {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("编码请求失败: %v", err)
		}
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("执行请求失败: %v", err)
	}
	return response
}

func decodeProxySiteResponse(t *testing.T, response *http.Response) SiteResponse {
	t.Helper()
	var result SiteResponse
	decodeProxySiteJSON(t, response, &result)
	return result
}

func decodeProxySiteJSON(t *testing.T, response *http.Response, target any) {
	t.Helper()
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
}
