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
| GET | `/api/proxy-sites?page=1&page_size=20` | 分页列表，默认每页 20 条 |
| POST | `/api/proxy-sites` | 新增 |
| GET | `/api/proxy-sites/{id}` | 详情 |
| PUT | `/api/proxy-sites/{id}` | 全量更新 |
| DELETE | `/api/proxy-sites/{id}` | 软删除 |
| POST | `/api/proxy-sites/{id}/clone` | 克隆，可覆盖域名和上游 |
| POST | `/api/proxy-sites/{id}/enable` | 启用 |
| POST | `/api/proxy-sites/{id}/disable` | 停用 |
| POST | `/api/proxy-sites/{id}/preview` | 已保存站点的 Caddy JSON 与 Caddyfile 预览 |
| POST | `/api/proxy-sites/preview` | 尚未保存站点草稿的 Caddy JSON 与 Caddyfile 预览 |
| POST | `/api/proxy-sites/import/nginx` | 导入 Nginx server/upstream 配置为停用站点 |

数组字段在 API 中使用 JSON 数组，在 SQLite 中编码为 JSON string。站点名称由首个域名兼容生成。克隆接口无论原站点状态如何都会生成 `enabled=false` 的新记录。

`site_type` 支持：

- `proxy`：传统反向代理，`upstream_type` 继续表示 HTTP、HTTPS、h2c 或 Unix Socket 协议。
- `static`：从 `root_path` 提供静态文件，不要求 `upstreams`。
- `spa`：`api_path` 匹配的请求反代到 `upstreams`，其余请求从 `root_path` 提供并 fallback 到 `index.html`。

静态和 SPA 模式可通过 `enable_security_headers` 与 `enable_asset_cache` 开启常用安全头和静态资源缓存策略。

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
| GET/POST | `/api/certificates` | 查询或新增可复用证书配置；列表包含站点引用数、签发状态、最近错误和实际证书信息 |
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
| POST | `/api/caddy/upload` | 上传可执行文件、ZIP 或 tar.gz 并切换托管 Caddy |
| GET | `/api/caddy/update-tasks/current` | 获取当前更新任务阶段、进度和失败原因 |
| GET | `/api/caddy/change-status` | 动态比较业务配置、持久化配置与 Caddy 当前 JSON |
| POST | `/api/caddy/preview` | 生成完整 JSON，不发布 |
| POST | `/api/caddy/preview-caddyfile` | 从结构化站点生成只读 Caddyfile，并调用托管 Caddy `adapt` 校验语法 |
| POST | `/api/caddy/validate` | 对生成配置执行 JSON 与管理入口基础校验 |
| POST | `/api/caddy/publish` | 生成版本并调用 `/load` |
| GET | `/api/caddy/current-config` | 调用 Caddy `GET /config/` |

配置版本详情可选返回 `caddyfile`。旧版本没有留存内容时省略该字段；发布和回滚始终使用 `caddy_json`。

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

## 日志

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/logs?source=system&cursor=0&limit=200` | 通过游标增量读取日志；`source` 也可使用 `caddy`、`access` 或脱敏后的 `dns` Provider 日志 |
| GET | `/api/logs?source=dns&provider_id={id}` | 读取指定 DNS Provider 的脱敏调用日志 |

## Dashboard

`GET /api/dashboard/summary` 返回：

```json
{
  "site_count": 3,
  "enabled_site_count": 2,
  "disabled_site_count": 1,
  "https_site_count": 2,
  "request_count_24h": 1280,
  "error_count_24h": 3,
  "traffic_bytes_24h": 73400320,
  "top_sites_24h": [
    { "domain": "app.example.com", "request_count": 900, "error_count": 2, "bytes": 52428800 }
  ],
  "last_publish_time": "2026-06-20T10:00:00+08:00",
  "caddy_online": true
}
```

## 错误处理

Huma 使用标准问题详情结构返回错误。发布或回滚中的 Caddy 调用失败通常返回 502，同时数据库内会保留 `failed` 版本及 `error_message`。
