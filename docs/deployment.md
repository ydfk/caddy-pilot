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
| `CADDY_BINARY` | `caddy` | 启动检查和版本读取使用的 Caddy 可执行文件 |
| `CADDY_VERSION_CHECK_URL` | GitHub Caddy latest release API | 版本校验地址，支持 GitHub 或 `{version, update_url}` JSON |
| `CADDY_UPDATE_URL` | 空 | 可选的自定义更新地址，设置后覆盖版本服务返回地址 |
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

管理界面的“Caddy 状态 → 版本管理”会显示当前版本与 GitHub 最新稳定版。CaddyPilot 不在运行中的容器内替换 Caddy 二进制，因为该修改在容器重建后会丢失，也可能中断管理入口。

后端启动时会通过 PATH 或 `CADDY_BINARY` 定位真实可执行文件，并执行 `caddy version`。生产环境检查失败会终止启动；开发环境会输出警告并继续，以便只调试管理界面。

固定或更新 Caddy 版本时使用镜像重建：

```powershell
git pull
$env:CADDY_VERSION = "2.10.0"
docker compose up -d --build
```

版本管理卡片可以根据最新稳定版生成并复制同类命令。镜像重建不会删除挂载的 `./data`。

## 故障排查

- 管理页不可访问：查看容器健康状态和 Caddy 日志。
- API 返回 401：重新登录；浏览器会自动清理失效 Token。
- 发布返回 502：检查 Caddy 状态页与失败版本的 `error_message`。
- 站点发布后不可达：检查域名解析、80/443 防火墙、上游地址和 Caddy 当前 JSON。
- 容器反复重启：任一关键进程退出都会触发整体重启，应查看重启前日志定位根因。

## 重启后的配置

由于 Caddy 设置 `persist_config off`，容器重启后会先加载镜像内固定 Caddyfile，确保管理界面可访问。业务代理站点需要在 Caddy 状态页再次发布。自动恢复最近成功版本列为后续 TODO。
