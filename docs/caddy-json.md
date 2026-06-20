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

请求 Header 写入 `reverse_proxy.headers.request.set`。多个上游写为多个 `dial`，由 Caddy 执行负载分配。WebSocket 无需额外配置，Caddy reverse_proxy 会处理协议升级。

## HTTPS 与证书

HTTPS server 使用 host matcher，Caddy 会补充 TLS connection policy 并执行自动证书管理。全局关闭 Caddy 自动 HTTP 跳转，由生成器按站点的 `force_https` 显式控制。

## 回滚修复

`EnsureManagementEntry` 会：

1. 解析历史 JSON。
2. 查找并删除任何监听 `:8080` 的冲突 server。
3. 注入受保护管理 server。
4. 强制 Admin API 回到 `127.0.0.1:2019`。

JSON 无法解析时回滚失败，并创建带错误信息的失败版本。

## 单元测试

测试至少覆盖空站点、host 与 reverse_proxy、停用过滤、多上游、管理入口注入和历史配置修复。还使用官方 Caddy 镜像对代表性生成结果执行过 `caddy validate`。
