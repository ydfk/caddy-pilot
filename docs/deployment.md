# 部署与运维

## 前置条件

- Docker Engine 及 Docker Compose v2
- 可用端口 80、443、8080
- 用于签发证书时，域名应解析到部署主机且 80/443 可从公网访问

## 启动

```powershell
docker compose up -d --build
docker compose ps
```

源码构建会主动拉取 `caddy:2-alpine` 与 `caddy:2-builder-alpine`，因此每次重新构建都会采用当时最新的稳定 Caddy 2.x。

健康状态变为 `healthy` 后访问 <http://localhost:8080>。首次使用选择“初始化管理员”。

默认 `docker-compose.yml` 可直接用于生产单机部署。用户只需要确认端口可用并持久化 `./data`，不需要声明容器内部地址、Admin API、数据库路径或密钥。

若生产服务器直接从 Docker Hub 拉取镜像，使用：

```powershell
docker compose -f docker-compose.prod.yml up -d
```

默认镜像为 `ydfk/caddy-pilot:latest`。可通过 `CADDYPILOT_VERSION` 固定版本，例如 `1.2.3`；生产环境建议固定版本后再升级。

## 端口

| 环境变量 | 默认宿主机端口 | 容器端口 | 用途 |
| --- | --- | --- | --- |
| `CADDYPILOT_HTTP_PORT` | 80 | 80 | HTTP 代理站点 |
| `CADDYPILOT_HTTPS_PORT` | 443 | 443 | HTTPS 代理站点及全局跳转目标 |
| — | 8080 | 8080 | CaddyPilot 管理界面 |

需要使用非标准宿主机端口时，在项目根目录创建 `.env`：

```dotenv
CADDYPILOT_HTTP_PORT=18080
CADDYPILOT_HTTPS_PORT=18443
```

随后访问 `http://主机:18080`，强制 HTTPS 会统一跳转到 `https://主机:18443`。不要使用 `10080`：Chromium 系浏览器将其列为不安全端口并直接返回 `ERR_UNSAFE_PORT`，Caddy 无法覆盖浏览器的限制。

2019 没有宿主机映射。**不要将 Caddy Admin API 2019 端口暴露到公网。** 使用 HTTP-01 签发证书时，公网仍需能访问标准 80 端口；非标准映射更适合 DNS-01 或由上层网关转发 80/443 的部署。

静态目录或 SPA 站点的 `root_path` 是容器内路径，需要在 Compose 中额外挂载只读目录，例如：

```yaml
volumes:
  - ./data:/data
  - /srv/www:/var/www:ro
```

## 自动管理的内部配置

镜像内部自动设置后端地址、Caddy Admin API、数据库路径、运行时目录和管理入口。首次启动还会生成两份独立随机密钥：

- JWT 签名密钥。
- DNS Provider 凭据加密密钥。

密钥保存到 `/data/.caddypilot-secrets`，文件权限为 `0600`。容器重建和镜像升级会继续使用原密钥，Compose 不需要也不应该重复声明这些内部配置。

版本校验地址和下载地址不参与容器正常启动：镜像构建时会使用带阿里云 DNS 审计模块的最新稳定 Caddy 2.x。国内网络无法访问 GitHub 时，只会影响“检查更新”和“在线更新”，不会影响已有站点运行。

这三个地址在“Caddy 管理 → 更新源设置”中修改并保存到 SQLite，不再使用 Docker 环境变量：

- 版本校验地址：默认读取 Caddy 官方 GitHub Release，也兼容返回 `tag_name` 或 `version` 的 JSON。
- Caddy 下载地址：默认使用 Caddy 官方动态构建服务并包含 AliDNS 模块，也支持 `{version}`、`{os}`、`{arch}`、`{ext}` 占位符。
- SHA-512 清单地址：官方动态构建源默认留空；自定义静态下载源可填写清单并在安装前强制校验。

在线下载支持断点续传和最多三次退避重试，页面刷新后仍会恢复任务进度或最后一次失败原因。普通部署无需修改；国内部署可配置可信镜像，或使用“上传 Caddy 安装包”作为兜底。

## 检查

```powershell
docker compose ps
docker compose logs --tail 100
curl.exe -I http://localhost:8080/
docker port caddypilot
```

`docker port caddypilot` 的输出不应出现 2019。

## 数据备份

停止写入后复制整个 `./data`：

```powershell
docker compose stop
Copy-Item -Recurse -LiteralPath .\data -Destination .\backup\data
docker compose start
```

恢复前应先停止容器，并确认备份版本与当前镜像兼容。备份必须包含 `.caddypilot-secrets`；丢失该文件后，数据库内已有的 DNS Provider 密文将无法解密。

## 更新

管理界面的“Caddy 状态 → 版本管理”会显示当前版本与最新稳定版。点击“后端更新”后，后端下载独立版本目录、保护当前管理配置、重启 Caddy，并在失败时尝试恢复旧版本。

用户不需要安装或启动 Caddy。Docker 镜像带有基础版本；Windows 独立环境缺少运行时时，后端会根据目标版本和系统中保存的下载地址自动下载到项目 `data/runtime/`。

Docker 镜像和默认托管下载都包含 `github.com/caddy-dns/alidns`。使用自定义下载地址时，必须确保目标 Caddy 同样包含该模块，否则阿里云 DNS-01 配置无法加载。

阿里云 AccessKey 在“证书与访问 → DNS Provider”中配置，系统使用自动生成的独立密钥进行 AES-GCM 加密。

Windows 无 Docker 开发使用：

```powershell
.\scripts\dev.cmd
```

完整管理界面统一访问 <http://localhost:8080>；Vite 的 3000 端口只是内部开发上游。

## Docker Hub 自动发布

GitHub Actions 仅响应 `v1.2.3` 或 `v1.2.3-rc.1` 格式的 Tag，并发布 `linux/amd64`、`linux/arm64` 镜像。仓库需要配置：

- `DOCKERHUB_USERNAME`：Docker Hub 用户名。
- `DOCKERHUB_TOKEN`：仅授予仓库写入权限的 Access Token。

稳定版本会生成完整版本、主次版本、主版本和 `latest` 标签；预发布版本不会覆盖 `latest`。

系统侧栏显示的 CaddyPilot 版本由同一个 Tag 写入。例如 `v1.2.3` 会发布 Docker 镜像 `1.2.3`，系统内显示版本同样为 `1.2.3`。

## 故障排查

- 管理页不可访问：查看容器健康状态和 Caddy 日志。
- API 返回 401：重新登录；浏览器会自动清理失效 Token。
- 发布返回 502：检查 Caddy 状态页与失败版本的 `error_message`。
- 站点发布后不可达：检查域名解析、80/443 防火墙、上游地址和 Caddy 当前 JSON。
- 容器反复重启：托管 Caddy 或后端异常退出都会触发整体重启，应查看重启前日志定位根因。

## 重启后的配置

由于 Caddy 关闭配置持久化，系统重启后会先加载后端生成的受保护管理配置，确保管理界面可访问。业务代理站点需要在 Caddy 状态页再次发布。自动恢复最近成功版本列为后续 TODO。
