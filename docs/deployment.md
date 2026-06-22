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

健康状态变为 `healthy` 后访问 <http://localhost:8080>。首次使用选择“初始化管理员”。

默认 `docker-compose.yml` 可直接用于生产单机部署。用户只需要确认端口可用并持久化 `./data`，不需要声明容器内部地址、Admin API、数据库路径或密钥。

## 端口

| 宿主机 | 容器 | 用途 |
| --- | --- | --- |
| 80 | 80 | HTTP 代理站点 |
| 443 | 443 | HTTPS 代理站点 |
| 8080 | 8080 | CaddyPilot 管理界面 |

2019 没有宿主机映射。**不要将 Caddy Admin API 2019 端口暴露到公网。**

## 自动管理的内部配置

镜像内部自动设置后端地址、Caddy Admin API、数据库路径、运行时目录和管理入口。首次启动还会生成两份独立随机密钥：

- JWT 签名密钥。
- DNS Provider 凭据加密密钥。

密钥保存到 `/data/.caddypilot-secrets`，文件权限为 `0600`。容器重建和镜像升级会继续使用原密钥，Compose 不需要也不应该重复声明这些内部配置。

只有开发、排障或接入自定义下载源时，才需要临时覆盖 `CADDY_VERSION_CHECK_URL`、`CADDY_DOWNLOAD_URL` 或 `CADDY_CHECKSUM_URL`。普通生产部署无需关注。

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

用户不需要安装或启动 Caddy。Docker 镜像带有基础版本；Windows 独立环境缺少运行时时，后端会根据 `CADDY_VERSION` 和 `CADDY_DOWNLOAD_URL` 自动下载到项目 `data/runtime/`。

Docker 镜像和默认托管下载都包含 `github.com/caddy-dns/alidns`。使用自定义 `CADDY_DOWNLOAD_URL` 时，必须确保目标 Caddy 同样包含该模块，否则阿里云 DNS-01 配置无法加载。

阿里云 AccessKey 在“证书与访问 → DNS Provider”中配置，系统使用自动生成的独立密钥进行 AES-GCM 加密。

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
