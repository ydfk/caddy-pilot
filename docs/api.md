# API 说明

## 约定

- 基础路径：同域 `/api`
- 数据格式：JSON
- 认证：`Authorization: Bearer <JWT>`
- OpenAPI 3.1：`/openapi.yaml`
- 交互文档：`/docs`

除初始化状态、首次注册与登录接口外，以下接口都需要 JWT。

## 认证

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/auth/setup-status` | 查询是否已创建管理员 |
| POST | `/api/auth/register` | 首次创建管理员 |
| POST | `/api/auth/login` | 登录并返回 JWT |
| GET | `/api/auth/profile` | 当前用户 |

## 代理站点

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/proxy-sites` | 列表 |
| POST | `/api/proxy-sites` | 新增 |
| GET | `/api/proxy-sites/{id}` | 详情 |
| PUT | `/api/proxy-sites/{id}` | 全量更新 |
| DELETE | `/api/proxy-sites/{id}` | 软删除 |
| POST | `/api/proxy-sites/{id}/clone` | 克隆，可覆盖域名和上游 |
| POST | `/api/proxy-sites/{id}/enable` | 启用 |
| POST | `/api/proxy-sites/{id}/disable` | 停用 |
| POST | `/api/proxy-sites/{id}/preview` | 站点路由 JSON 片段 |

数组字段在 API 中使用 JSON 数组，在 SQLite 中编码为 JSON string。站点名称由首个域名兼容生成。克隆接口无论原站点状态如何都会生成 `enabled=false` 的新记录。

## Basic Auth 密码本

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/basic-auth-credentials` | 密码条目列表，不返回密码或哈希 |
| POST | `/api/basic-auth-credentials` | 新增条目并生成 bcrypt 哈希 |
| PUT | `/api/basic-auth-credentials/{id}` | 修改名称、用户名或密码 |
| DELETE | `/api/basic-auth-credentials/{id}` | 删除未被站点引用的条目 |

## DNS Provider 与证书

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET/POST | `/api/dns-providers` | 查询或新增系统 DNS Provider |
| PUT/DELETE | `/api/dns-providers/{id}` | 编辑或删除未被引用的 Provider |
| GET/POST | `/api/certificates` | 查询或新增可复用证书配置 |
| PUT/DELETE | `/api/certificates/{id}` | 编辑或删除未被站点引用的证书配置 |

## Caddy 工作台设置

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/caddy/settings` | 获取数据库中保存的版本校验、下载与校验和地址 |
| PUT | `/api/caddy/settings` | 更新 Caddy 更新源设置 |

## 系统信息

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/system/info` | 获取独立于 Caddy 运行时的 CaddyPilot 系统版本 |

## Caddy

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/caddy/status` | Caddy 在线状态 |
| GET | `/api/caddy/version` | 当前版本、最新稳定版与托管下载信息 |
| POST | `/api/caddy/update` | 由后端异步下载、切换并重启托管 Caddy |
| GET | `/api/caddy/change-status` | 当前启用站点是否存在未发布变更 |
| POST | `/api/caddy/preview` | 生成完整 JSON，不发布 |
| POST | `/api/caddy/validate` | 对生成配置执行 JSON 与管理入口基础校验 |
| POST | `/api/caddy/publish` | 生成版本并调用 `/load` |
| GET | `/api/caddy/current-config` | 调用 Caddy `GET /config/` |

发布请求体：

```json
{
  "reason": "手动发布"
}
```

## 配置版本

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/config-versions` | 按版本号倒序列表 |
| GET | `/api/config-versions/{id}` | 业务快照、Caddy JSON 和错误详情 |
| POST | `/api/config-versions/{id}/rollback` | 创建新版本并回滚 |

状态值：`draft`、`published`、`failed`、`rollback`。

## Dashboard

`GET /api/dashboard/summary` 返回：

```json
{
  "site_count": 3,
  "enabled_site_count": 2,
  "disabled_site_count": 1,
  "https_site_count": 2,
  "last_publish_time": "2026-06-20T10:00:00+08:00",
  "caddy_online": true,
  "caddy_admin_api": "http://127.0.0.1:2019"
}
```

## 错误处理

Huma 使用标准问题详情结构返回错误。发布或回滚中的 Caddy 调用失败通常返回 502，同时数据库内会保留 `failed` 版本及 `error_message`。
