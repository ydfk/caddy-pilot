# Caddy JSON 生成规则

## 输入与输出

`backend/internal/caddygen` 接收 `ProxySite` 列表，忽略软删除和 `enabled=false` 站点，输出可以直接提交给 Caddy `/load` 的完整 JSON。

顶层始终包含：

```json
{
  "admin": {
    "listen": "127.0.0.1:2019",
    "config": { "persist": false }
  },
  "apps": {
    "http": { "servers": {} }
  }
}
```

## 管理服务器

固定 server 名为 `caddypilot-admin`，监听 `:8080`：

- `/api/*` 原样反向代理到 `127.0.0.1:25610`
- 其他路径从 `/app/frontend` 提供文件
- `try_files` 回退 `/index.html`
- gzip 与 zstd 编码

发布和回滚都不允许删除该 server。

## 业务服务器

- `proxy-http` 监听 `:80`
- `proxy-https` 监听 `:443`
- 开启 HTTPS 的站点进入 HTTPS server
- 同时开启 Force HTTPS 的站点在 HTTP server 生成 308 跳转
- 未开启 Force HTTPS 的站点继续在 HTTP server 提供代理

每个站点生成 host matcher 和 subroute。IP 白名单存在时，会在同一 matcher 中追加 `remote_ip`，与 host 条件共同生效。

## Handler 顺序

站点 subroute 按需生成：

1. 响应 Header handler
2. gzip/zstd encode handler
3. Basic Auth authentication handler
4. reverse_proxy handler

请求 Header 写入 `reverse_proxy.headers.request.set`。多个上游写为多个 `dial`，由 Caddy 执行负载分配。HTTP 使用默认 transport；HTTPS 增加 TLS、SNI 和可选的跳过校验；h2c 使用 `versions: ["h2c"]`；Unix Socket 使用 `unix/` 网络地址。WebSocket 无需额外配置。

## HTTPS 与证书

HTTPS server 显式生成 TLS connection policy。域名证书默认使用 HTTP/TLS-ALPN 验证，也可以选择 DNS-01；通配符证书引用系统证书配置并强制使用 DNS-01。阿里云凭据在数据库中加密保存，生成 JSON 只写入按 Provider ID 隔离的环境变量占位符，因此 AccessKey 不会进入配置版本。全局关闭 Caddy 自动 HTTP 跳转，由生成器按站点的 `force_https` 显式控制。

## 回滚修复

`EnsureManagementEntry` 会：

1. 解析历史 JSON。
2. 查找并删除任何监听 `:8080` 的冲突 server。
3. 注入受保护管理 server。
4. 强制 Admin API 回到 `127.0.0.1:2019`。

JSON 无法解析时回滚失败，并创建带错误信息的失败版本。

## 单元测试

测试覆盖空站点、host 与 reverse_proxy、停用过滤、多上游、类型化 transport、TLS policy、阿里云 DNS-01、管理入口注入和历史配置修复。
