# CaddyPilot

## 项目介绍

CaddyPilot 是一个轻量、单节点的 Caddy 反向代理可视化管理工具。它把 React 管理界面、Go API 和 Caddy 放进同一个 Docker 镜像，通过受保护的 `:8080` 管理入口完成站点维护、配置预览、发布、版本记录与回滚。

## 功能范围

- 单用户 JWT 登录与首次管理员初始化
- 代理站点新增、编辑、软删除、克隆、启用和停用
- 克隆站点默认停用
- 多域名、多上游、HTTPS 跳转、压缩、Header、IP 白名单和 Basic Auth 配置
- 单站点 JSON 片段与完整 Caddy JSON 预览
- 发布配置到本机 Caddy Admin API
- 发布失败留痕、配置版本列表、详情和受保护回滚
- Dashboard 与 Caddy 当前配置查看
- 单镜像 Docker Compose 部署

## 不做什么

MVP 不支持多 Caddy 节点、多用户权限、插件市场、Caddyfile 导入导出、DNS Provider 管理、Layer4、Docker 容器自动发现、复杂日志分析和完整 Caddy 配置可视化。

## 架构说明

容器内由 supervisor 同时管理 Caddy 与 Go 后端：

```text
浏览器 -> :8080 Caddy -> /api/* -> 127.0.0.1:25610 Go API
                    \-> React 静态文件 /app/frontend

Go API -> SQLite /data/caddypilot.db
Go API -> Caddy Admin API 127.0.0.1:2019
```

每次发布和回滚都会检查并注入等效的 `:8080` 管理服务器，避免新配置切断 CaddyPilot 自身入口。详细设计见 [docs/design.md](docs/design.md)。

## 单镜像部署

镜像包含：

- Caddy 2.10
- CaddyPilot Go 后端
- React 生产静态文件
- supervisor 与关键进程退出监听器
- CA 证书与时区数据

后端使用 CGO 构建官方 GORM SQLite 驱动。任一关键进程退出时 supervisor 会结束，Compose 的重启策略随后重启整个容器。

## Docker Compose 使用方法

建议先设置独立 JWT 密钥：

```powershell
$env:JWT_SECRET = "请替换为足够长的随机字符串"
docker compose up -d --build
docker compose ps
docker compose logs -f
```

停止服务：

```powershell
docker compose down
```

## 默认访问地址

管理界面默认地址：<http://localhost:8080>。首次启动且没有用户数据时，登录页会自动显示初始化界面；创建唯一管理员后，初始化入口会关闭。

## Caddy Admin API 安全提醒

**不要将 Caddy Admin API 2019 端口暴露到公网。**

默认 Compose 没有映射 2019。Admin API 只监听容器内的 `127.0.0.1:2019`，管理界面也不会将它代理给浏览器。

## 数据目录

Compose 将根目录的 `./data` 挂载到容器 `/data`：

- `/data/caddypilot.db`：用户、代理站点和配置版本
- `/data/caddy`：Caddy 证书与运行数据

备份时应同时保存 SQLite 文件和 Caddy 数据目录。`data/` 已加入 Git 忽略列表。

## 开发环境启动

Windows 可直接双击根目录的 `dev.cmd`，或在 PowerShell 中一键启动：

```powershell
.\dev.cmd
```

脚本只启动本机 Go 后端与 Vite 前端，不使用 Docker。它会自动安装前端依赖；按 `Ctrl+C` 可同时停止两个进程。管理界面地址为 <http://localhost:3000>，API 会代理到 `http://127.0.0.1:25610`。

仅检查 Go 与 pnpm 启动环境：

```powershell
.\dev.cmd -Check
```

如果需要分别启动，也可以使用：

后端：

```powershell
cd backend
go test ./...
go run ./cmd
```

前端：

```powershell
cd frontend
pnpm install
pnpm dev
```

Vite 默认在 `http://localhost:3000` 启动，并将 `/api` 原样代理到 `http://127.0.0.1:25610`。

## 文档

- [设计与安全边界](docs/design.md)
- [API 说明](docs/api.md)
- [Caddy JSON 生成规则](docs/caddy-json.md)
- [部署与运维](docs/deployment.md)

## 后续计划

- TODO：将 `advanced_json` 按受控白名单合并到生成配置；当前仅保存
- TODO：提供 Basic Auth bcrypt 哈希生成助手；当前表单要求输入哈希
- TODO：把 `enable_log` 从已保存开关扩展为精细的站点级访问日志配置
- TODO：支持容器重启后按策略自动恢复最近成功版本；当前重启后由初始 Caddyfile 保证管理入口可用
