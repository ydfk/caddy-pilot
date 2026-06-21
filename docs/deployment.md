# 部署与运维

## 前置条件

- Docker Engine 及 Docker Compose v2
- 可用端口 80、443、8080
- 用于签发证书时，域名应解析到部署主机且 80/443 可从公网访问

## 启动

```powershell
$env:JWT_SECRET = "请替换为随机密钥"
docker compose up -d --build
docker compose ps
```

健康状态变为 `healthy` 后访问 <http://localhost:8080>。首次使用选择“初始化管理员”。

## 端口

| 宿主机 | 容器 | 用途 |
| --- | --- | --- |
| 80 | 80 | HTTP 代理站点 |
| 443 | 443 | HTTPS 代理站点 |
| 8080 | 8080 | CaddyPilot 管理界面 |

2019 没有宿主机映射。**不要将 Caddy Admin API 2019 端口暴露到公网。**

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `TZ` | `Asia/Shanghai` | 容器时区 |
| `APP_ENV` | `production` | 后端运行环境 |
| `JWT_SECRET` | Compose 开发默认值 | 生产环境必须覆盖 |
| `DATABASE_DSN` | `/data/caddypilot.db` | SQLite 路径 |
| `CADDY_ADMIN_API` | `http://127.0.0.1:2019` | 容器内 Admin API |
| `CADDY_VERSION` | `2.10.0` | 后端托管的目标版本 |
| `CADDY_VERSION_CHECK_URL` | GitHub Caddy latest release API | 版本校验地址，支持 GitHub 或 `{version, update_url}` JSON |
| `CADDY_DOWNLOAD_URL` | GitHub release asset 模板 | 二进制下载地址，支持 `{version}`、`{os}`、`{arch}`、`{ext}` 占位符 |
| `CADDY_CHECKSUM_URL` | GitHub release checksum 模板 | 下载后用于 SHA-512 完整性校验 |
| `CADDYPILOT_RUNTIME_DIR` | `/data/runtime` | 后端托管 Caddy 版本和启动配置目录 |
| `CADDY_DATA_DIR` | `/data/caddy` | Caddy 证书与运行数据目录 |
| `CADDYPILOT_BACKEND_ADDR` | `127.0.0.1:25610` | 后端监听地址 |
| `CADDYPILOT_FRONTEND_DIR` | `/app/frontend` | 前端静态目录 |
| `CADDYPILOT_MANAGE_ADDR` | `:8080` | 管理入口标识 |

## 检查

```powershell
docker compose ps
docker compose logs --tail 100
curl.exe -I http://localhost:8080/
docker port caddypilot
```

`docker port caddypilot` 的输出不应出现 2019。

## 数据备份

停止写入后复制 `./data`：

```powershell
docker compose stop
Copy-Item -Recurse -LiteralPath .\data -Destination .\backup\data
docker compose start
```

恢复前应先停止容器，并确认备份版本与当前镜像兼容。

## 更新

管理界面的“Caddy 状态 → 版本管理”会显示当前版本与最新稳定版。点击“后端更新”后，后端下载独立版本目录、保护当前管理配置、重启 Caddy，并在失败时尝试恢复旧版本。

用户不需要安装或启动 Caddy。Docker 镜像带有基础版本；Windows 独立环境缺少运行时时，后端会根据 `CADDY_VERSION` 和 `CADDY_DOWNLOAD_URL` 自动下载到项目 `data/runtime/`。

Windows 无 Docker 开发使用：

```powershell
.\dev.cmd
```

完整管理界面统一访问 <http://localhost:8080>；Vite 的 3000 端口只是内部开发上游。

## 故障排查

- 管理页不可访问：查看容器健康状态和 Caddy 日志。
- API 返回 401：重新登录；浏览器会自动清理失效 Token。
- 发布返回 502：检查 Caddy 状态页与失败版本的 `error_message`。
- 站点发布后不可达：检查域名解析、80/443 防火墙、上游地址和 Caddy 当前 JSON。
- 容器反复重启：托管 Caddy 或后端异常退出都会触发整体重启，应查看重启前日志定位根因。

## 重启后的配置

由于 Caddy 关闭配置持久化，系统重启后会先加载后端生成的受保护管理配置，确保管理界面可访问。业务代理站点需要在 Caddy 状态页再次发布。自动恢复最近成功版本列为后续 TODO。
