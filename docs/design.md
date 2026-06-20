# 设计与安全边界

## 目标

CaddyPilot 面向单机、自托管场景，以尽量小的运维面完成反向代理站点的可视化管理。系统坚持单 Caddy 节点、单管理员和 SQLite，不把 MVP 扩展为通用 Caddy 控制平面。

## 组件

| 组件 | 容器内地址 | 职责 |
| --- | --- | --- |
| Caddy | `:80`、`:443`、`:8080` | 用户代理流量、管理界面和静态文件 |
| Caddy Admin API | `127.0.0.1:2019` | 接收配置与返回当前 JSON |
| Go API | `127.0.0.1:25610` | 认证、业务数据、生成、发布和回滚 |
| React | `/app/frontend` | 管理界面静态文件 |
| SQLite | `/data/caddypilot.db` | 持久化业务数据和版本历史 |

## 固定管理入口

管理入口是系统最重要的不变量：

1. 必须存在监听 `:8080` 的 HTTP server。
2. `/api/*` 必须保持原路径代理到 `127.0.0.1:25610`。
3. 其他路径使用 `/app/frontend`，不存在的静态路径回退到 `/index.html`。
4. 启用 gzip 与 zstd。

计划示例使用了 `handle_path /api/*`，但该指令会剥离 `/api` 前缀，与后端实际 `/api/...` 路由冲突。因此实现使用 `handle /api/*`，保持功能目标不变并避免接口 404。

`caddygen.Generate` 总是注入管理服务器。`EnsureManagementEntry` 在回滚前移除冲突的 `:8080` server，再注入受保护版本。发布前还会使用 `HasManagementEntry` 进行最终检查。

## 发布状态机

```text
读取 enabled 站点
  -> 生成业务快照和 Caddy JSON
  -> 创建 draft ConfigVersion
  -> 检查管理入口
  -> POST /load
     -> 成功：published + published_at
     -> 失败：failed + error_message
```

失败记录不会覆盖或删除之前的成功版本。

## 回滚状态机

```text
读取目标历史版本
  -> 解析历史 Caddy JSON
  -> 强制注入管理入口
  -> 创建新的 draft 记录
  -> POST /load
     -> 成功：新记录标记 rollback
     -> 失败：新记录标记 failed
```

回滚不会修改历史行，因而版本记录保持可审计。

## 认证边界

- 登录后使用 HS256 JWT。
- 除注册和登录外，所有业务 API 都声明 Bearer Auth。
- 浏览器端遇到 401 会清除本地 Token 并返回登录页。
- 系统定位为单管理员；首次初始化后不应继续创建其他管理员。

## 进程模型

supervisor 运行 Caddy、Go 后端和退出监听器。监听器订阅关键进程的 `PROCESS_STATE_EXITED` 与 `PROCESS_STATE_FATAL`，任一关键进程退出都会终止 supervisor，使容器由 Compose 整体重启。

## 当前限制

- `advanced_json` 仅保存，不参与合并。
- WebSocket 依赖 Caddy `reverse_proxy` 的原生协议升级能力，不生成额外 handler。
- Caddy 配置设置 `persist_config off`；容器重启后先加载固定管理入口，需手动再次发布业务站点。
