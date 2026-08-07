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
- 密码和 Passkey 登录最终签发相同的 JWT；密码始终保留为恢复方式。
- Passkey 注册、列表、重命名和删除要求 Bearer Auth；Passkey 登录挑战与验证允许未登录访问。
- WebAuthn RP ID 与允许 Origin 由配置或环境变量提供，凭据内容加密落库，挑战保存在服务端且短期、一次性有效。
- 除初始化、密码登录、Passkey 登录和公开状态检查外，所有业务 API 都声明 Bearer Auth。
- 浏览器端遇到 401 会清除本地 Token 并返回登录页。
- 系统定位为单管理员；首次初始化后不应继续创建其他管理员。

## 进程模型

Go 后端是系统主进程，负责准备、启动、监控、更新和关闭 Caddy。Docker 与 Windows 独立开发使用同一套托管逻辑：优先使用已选择的托管版本或镜像内置版本，缺失时下载到私有运行目录。Caddy 意外退出会触发后端整体退出，由 Compose 或上层服务管理器重启完整系统。

初始 Caddy JSON 由后端生成，不依赖静态 Caddyfile。生产环境从 `/app/frontend` 提供静态文件；本地开发环境把前端流量代理到 Vite。两种模式都保留相同的 `:8080` 管理入口与 `127.0.0.1:2019` Admin API 安全边界。

## 当前限制

- `advanced_json` 仅保存，不参与合并。
- WebSocket 依赖 Caddy `reverse_proxy` 的原生协议升级能力，不生成额外 handler。
- Caddy 配置关闭持久化；系统重启后先加载后端生成的固定管理入口，需手动再次发布业务站点。
