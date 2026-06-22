你现在需要完整创建并实现一个项目，项目名为 CaddyPilot。

CaddyPilot 是一个用于可视化管理 Caddy 反向代理配置的轻量 Web 工具。

## 实施进度

| 阶段 | 内容 | 状态 |
| --- | --- | --- |
| 1 | 初始化项目与模板基线 | 已完成 |
| 2 | 后端数据模型 | 已完成 |
| 3 | 代理站点 API | 已完成 |
| 4 | Caddy JSON 生成器及测试 | 已完成 |
| 5 | Caddy Admin API Client | 已完成 |
| 6 | 发布、校验、历史、回滚与 Dashboard API | 已完成 |
| 7 | 前端 API、认证、路由与布局 | 已完成 |
| 8 | 代理站点前端页面 | 已完成 |
| 9 | 配置版本前端页面 | 已完成 |
| 10 | Dashboard、Caddy 状态与设置页面 | 已完成 |
| 11 | 单镜像 Docker 部署 | 已完成 |
| 12 | 中文文档 | 已完成 |
| 13 | 全量质量与集成验收 | 已完成 |

请严格按照下面要求执行，不要跳过步骤。

### MVP 后续 TODO

- `advanced_json` 当前只保存，不合并到生成配置。
- Basic Auth 已改为统一密码本，后端自动生成 bcrypt 哈希且不会通过 API 回传密码或哈希。
- `enable_log` 当前只保存开关，后续补充精细的站点级访问日志配置。
- 容器重启后先恢复固定管理入口，后续补充最近成功版本的自动恢复策略。

### 2026-06-21 体验迭代

- [x] 收紧登录页、管理页、表单和卡片的空间密度。
- [x] 将站点状态、HTTPS、WebSocket、压缩、日志和认证模式改为选择控件。
- [x] 在代理站点页加入“校验配置 → 发布”引导操作。
- [x] 增加 Caddy 当前版本、最新稳定版检查与固定版本重建更新指引。
- [x] 仪表盘移除新增入口，站点页仅在存在未发布变更时引导校验与发布。
- [x] 站点编辑器集中核心字段，并将单选项改为开关或平铺单选。
- [x] 启动时真实检查 Caddy，并支持自定义版本校验地址与更新地址。
- [x] 将品牌 Logo 应用到登录页、侧栏和浏览器图标。
- [x] Caddy 改由 Go 后端统一准备、启动、监控、更新和关闭。
- [x] Docker 移除 supervisor/Caddyfile 双进程入口，后端作为唯一主进程托管 Caddy。
- [x] Windows 独立开发无需安装 Caddy，缺失时自动下载私有运行时并统一从 `:8080` 访问。

### 2026-06-21 单机管理深化

- [x] 移除多节点、本地控制平面、内部 Admin API 地址等界面信息。
- [x] 代理站点以域名和上游为核心，名称与描述退出编辑流程并由首个域名兼容生成。
- [x] 按 HTTP、HTTPS、h2c、Unix Socket 细化上游配置和 Caddy JSON。
- [x] 增加统一 Basic Auth 密码本，站点只引用密码条目。
- [x] 增加阿里云 DNS-01，并区分单域名与通配符证书。

### 2026-06-22 证书与系统配置

- [x] DNS Provider 改为系统级配置，凭据使用 AES-GCM 加密保存。
- [x] 增加可复用证书配置，站点引用系统证书。
- [x] 重组 Caddy 管理、证书与访问二级菜单。
- [x] 域名和上游改为明确的可重复输入项，并说明多值语义。
- [x] 将枚举选择收敛为紧凑下拉，仅保留布尔开关。
- [x] 完成后端、前端、单镜像 Docker 与浏览器交互验收。

### 2026-06-22 生产部署收敛

- [x] 默认 Compose 移除容器内部环境变量，仅保留端口、数据卷和运行策略。
- [x] 首次启动自动生成 JWT 与凭据加密密钥，并随 `/data` 持久化。
- [x] 验证生产镜像重建、健康检查和密钥跨重启保持不变。

### 2026-06-22 发布与脚本整理

- [x] 增加拉取 Docker Hub 镜像的 `docker-compose.prod.yml`。
- [x] 明确国内环境的版本检查与 Caddy 下载地址为可选更新能力。
- [x] 将跨模块开发、构建和测试脚本统一归档到根目录 `scripts/`。
- [x] 增加语义化 Tag 触发的多架构 Docker Hub 发布工作流。

### 2026-06-22 Caddy 工作台整合

- [x] 站点编辑器支持就地新增 DNS Provider 和 Basic Auth 密码条目。
- [x] 删除无业务设置项的系统设置菜单与页面。
- [x] 增加 CaddyPilot 系统版本，并通过镜像构建参数写入 Tag。
- [x] Caddy 更新源改为系统数据库配置，不再依赖 Docker 环境变量。
- [x] 将校验与发布整理为连续两步流程，仅在存在修改时展示。
- [x] 将 Caddy 运行状态、更新管理和配置版本合并为统一工作台。

### 2026-06-22 版本归属与运行时布局修正

- [x] CaddyPilot 系统版本移出 Caddy 工作台，通过独立系统接口在全局侧栏展示。
- [x] Tag、系统显示版本与 Docker 镜像版本统一使用去除 `v` 前缀后的版本号。
- [x] Caddy 在线状态、当前版本、最新版本和维护操作合并为运行时卡片。
- [x] 刷新与当前 JSON 移入运行时操作区，页面标题区不再堆放按钮。
- [x] 更新源入口改为明确的文字按钮，并补充 SHA-512 完整性校验说明。

### 2026-06-22 Caddy 持久化、证书与日志完善

- [ ] 发布构建仅保留 linux/amd64，移除 QEMU、SBOM 和高开销 provenance，并强化 BuildKit 缓存。
- [ ] Caddy 二进制、当前版本、活动配置、原生数据与日志统一持久化到 `/data`。
- [ ] 修复慢速下载超时，增加可轮询的更新任务状态与失败原因。
- [ ] 支持上传 Caddy 可执行文件、ZIP 和 tar.gz，并校验平台、版本及阿里云 DNS 模块。
- [ ] 启动恢复活动配置，发布与回滚原子持久化，避免站点配置在重启后丢失。
- [ ] 动态比较业务配置、持久化配置与 Caddy 当前 JSON，识别运行版本及配置漂移。
- [ ] 合并配置发布与配置版本界面，并提供漂移修复操作。
- [ ] 修复新增 DNS Provider 导致证书草稿丢失，所有普通弹窗禁止点击遮罩关闭。
- [ ] 统一 API 错误解析和表单内错误展示，覆盖证书、DNS、密码本、站点、更新、发布与回滚。
- [ ] 通配符证书允许多个域名且共用一个 DNS Provider，下拉仅显示配置名称。
- [ ] 移除无效 WebSocket 开关，保留旧字段兼容并固定新表单为 false。
- [ ] 从 Caddy PEM 证书读取域名、签发时间、到期时间、颁发机构、序列号与有效状态。
- [ ] 增加日志菜单，在线轮询查看系统日志与 Caddy 日志，日志文件写入 `/data/logs` 并轮转。
- [ ] 完成后端、前端、Action、单镜像 Docker 与浏览器全量验收。

> 2026-06-21 托管架构更新优先于下文早期的 supervisor、静态 Caddyfile 和手动启动描述；这些旧章节仅保留原始规划背景。
>
> 2026-06-21 单机管理深化同样优先于下文早期的 CaddyNode、站点名称/描述、站点内 Basic Auth JSON 和“不做 DNS Provider”描述；当前只管理本系统托管的 Caddy，并内置阿里云 DNS-01。

---

# 一、初始化项目

创建项目目录：

```bash
mkdir caddypilot
cd caddypilot
```

后端模板：

```text
https://github.com/ydfk/go-fiber-starter
```

前端模板：

```text
https://github.com/ydfk/react-starter
```

初始化命令：

```bash
git clone --depth=1 https://github.com/ydfk/go-fiber-starter backend
git clone --depth=1 https://github.com/ydfk/react-starter frontend

rm -rf backend/.git
rm -rf frontend/.git

git init
```

要求：

1. 不要保留两个模板仓库原有 Git 历史。
2. 后端目录必须叫 backend。
3. 前端目录必须叫 frontend。
4. 根目录必须包含 plan.md。
5. 先创建 plan.md，再开始写业务代码。
6. plan.md 内容按照本提示词中的设计整理。
7. 每完成一个主要阶段，更新 plan.md 的进度状态。

---

# 二、项目目标

MVP 目标：

1. 管理 Caddy 反向代理站点。
2. 支持新增、编辑、删除代理站点。
3. 支持克隆代理站点。
4. 支持启用 / 停用站点。
5. 支持生成 Caddy JSON。
6. 支持预览 Caddy JSON。
7. 支持发布配置到 Caddy Admin API。
8. 支持配置历史。
9. 支持配置回滚。
10. 支持单 Docker 镜像部署。
11. 前端、后端、Caddy 必须在同一个镜像里。
12. Docker 启动后通过 http://localhost:8080 访问管理界面。

MVP 不做：

1. 不做多 Caddy 节点。
2. 不做多用户权限。
3. 不做插件市场。
4. 不做 Caddyfile 导入。
5. 不做 Caddyfile 导出。
6. 不做阿里云以外的 DNS Provider 管理。
7. 不做 Layer4。
8. 不做 Docker 容器自动发现。
9. 不做复杂日志分析。
10. 不做完整 Caddy 全配置可视化。

---

# 三、单镜像架构

最终 Docker 镜像中必须包含：

1. Caddy
2. Go 后端
3. React 前端静态文件
4. supervisord 或等价进程管理工具

容器内部服务：

```text
Caddy HTTP: 80
Caddy HTTPS: 443
Caddy 管理界面入口: 8080
Caddy Admin API: 127.0.0.1:2019
Go Backend: 127.0.0.1:25610
React 前端目录: /app/frontend
SQLite 数据库: /data/caddypilot.db
```

容器启动后同时运行：

```text
1. caddy run --config /etc/caddy/Caddyfile --adapter caddyfile
2. /app/backend/caddypilot
```

使用 supervisord 管理两个进程。

要求：

1. 日志输出到 stdout / stderr。
2. 任意关键进程退出时容器应该退出。
3. 不要把 Caddy Admin API 暴露给宿主机。
4. docker-compose.yml 不允许映射 2019 端口。

---

# 四、最重要的架构约束

CaddyPilot 自己依赖 Caddy 提供 Web UI。

因此任何通过 CaddyPilot 发布的新 Caddy 配置，都必须保留 CaddyPilot 自身管理入口。

固定管理入口：

```text
:8080
```

职责：

1. 提供 React 前端静态文件。
2. 将 /api/* 反向代理到 127.0.0.1:25610。
3. 支持 React Router history fallback。
4. 启用 gzip / zstd 压缩。

逻辑 Caddyfile：

```text
:8080 {
    root * /app/frontend
    encode gzip zstd

    handle_path /api/* {
        reverse_proxy 127.0.0.1:25610
    }

    handle {
        try_files {path} /index.html
        file_server
    }
}
```

要求：

1. Caddy JSON 生成器必须自动注入等效的管理入口配置。
2. 用户不能通过 UI 删除这个入口。
3. 发布配置时必须检查最终 Caddy JSON 是否包含管理入口。
4. 回滚配置时也必须检查并注入管理入口。
5. 发布失败不能导致管理界面失联。
6. 回滚失败不能导致管理界面失联。

---

# 五、后端实现要求

后端使用 backend 模板原有架构。

技术要求：

1. Go
2. Fiber
3. Huma
4. GORM
5. JWT
6. SQLite
7. OpenAPI

必须保留模板已有认证能力。

所有新增业务接口都必须需要 JWT 认证。

---

# 六、后端模型

新增模型：

## ProxySite

字段：

```text
ID
Name
Description
Domains
Upstreams
EnableHTTPS
ForceHTTPS
EnableGzip
EnableLog
EnableWS
RequestHeaders
ResponseHeaders
BasicAuthEnabled
BasicAuthUsers
AllowedIPs
AdvancedJSON
Enabled
CreatedAt
UpdatedAt
DeletedAt
```

说明：

1. Domains 用 JSON string 存储。
2. Upstreams 用 JSON string 存储。
3. RequestHeaders 用 JSON string 存储。
4. ResponseHeaders 用 JSON string 存储。
5. BasicAuthUsers 用 JSON string 存储。
6. AllowedIPs 用 JSON string 存储。
7. AdvancedJSON MVP 只保存，不合并。

---

## ConfigVersion

字段：

```text
ID
Version
Reason
BusinessConfig
CaddyJSON
Status
ErrorMessage
PublishedAt
CreatedAt
```

Status 可选：

```text
draft
published
failed
rollback
```

---

## CaddyNode

字段：

```text
ID
Name
AdminAPI
Enabled
CreatedAt
UpdatedAt
```

默认创建：

```text
Name=local
AdminAPI=http://127.0.0.1:2019
Enabled=true
```

---

# 七、后端 API

## ProxySite API

```text
GET    /api/proxy-sites
POST   /api/proxy-sites
GET    /api/proxy-sites/{id}
PUT    /api/proxy-sites/{id}
DELETE /api/proxy-sites/{id}
POST   /api/proxy-sites/{id}/clone
POST   /api/proxy-sites/{id}/enable
POST   /api/proxy-sites/{id}/disable
POST   /api/proxy-sites/{id}/preview
```

要求：

1. DELETE 使用软删除。
2. clone 接口支持覆盖 name、domains、upstreams。
3. clone 后默认 enabled=false。
4. preview 返回当前站点生成的 Caddy JSON 片段。

---

## Caddy API

```text
GET  /api/caddy/status
POST /api/caddy/preview
POST /api/caddy/validate
POST /api/caddy/publish
GET  /api/caddy/current-config
```

要求：

1. preview 生成完整 Caddy JSON，但不发布。
2. validate 校验生成配置是否基本合法。
3. publish 调用 Caddy Admin API /load。
4. current-config 调用 Caddy Admin API /config/。

---

## ConfigVersion API

```text
GET  /api/config-versions
GET  /api/config-versions/{id}
POST /api/config-versions/{id}/rollback
```

---

## Dashboard API

```text
GET /api/dashboard/summary
```

返回：

```text
site_count
enabled_site_count
disabled_site_count
https_site_count
last_publish_time
caddy_online
caddy_admin_api
```

---

# 八、Caddy JSON 生成器

新增包：

```text
backend/internal/caddygen
```

功能：

1. 输入 ProxySite 列表。
2. 输出完整 Caddy JSON。
3. 只处理 enabled=true 的站点。
4. 支持 host matcher。
5. 支持 reverse_proxy。
6. 支持多个 upstream。
7. 支持 encode gzip / zstd。
8. 支持 request headers。
9. 支持 response headers。
10. 自动注入 CaddyPilot 管理入口。
11. advanced_json 暂时只保存，不合并。

必须有单元测试。

测试至少覆盖：

1. 空站点也能生成包含管理入口的 Caddy JSON。
2. 一个站点可以生成 host matcher 和 reverse_proxy。
3. disabled 站点不会进入 Caddy JSON。
4. 多 upstream 可以正确生成。
5. 生成结果必须包含 :8080 管理入口。

---

# 九、Caddy Admin API Client

新增：

```text
backend/internal/service/caddy_client.go
```

方法：

```text
GetConfig(ctx)
LoadConfig(ctx, config)
GetStatus(ctx)
```

环境变量：

```text
CADDY_ADMIN_API=http://127.0.0.1:2019
```

默认值：

```text
http://127.0.0.1:2019
```

要求：

1. HTTP 超时 5 秒。
2. LoadConfig 使用 POST /load。
3. GetConfig 使用 GET /config/。
4. GetStatus 可以使用 GET /config/ 判断。
5. 失败时返回清晰错误信息。

---

# 十、发布和回滚

发布流程：

```text
读取所有 enabled ProxySite
生成完整 Caddy JSON
自动注入 CaddyPilot 管理入口
保存 ConfigVersion draft
调用 Caddy Admin API /load
成功后标记为 published
失败后标记为 failed 并保存错误信息
```

回滚流程：

```text
读取历史 ConfigVersion
读取历史 caddy_json
检查是否包含 CaddyPilot 管理入口
如果没有，自动注入
调用 Caddy Admin API /load
成功后创建新的 ConfigVersion，status=rollback
```

要求：

1. 发布失败不能删除历史成功版本。
2. 回滚失败必须保存错误信息。
3. 发布和回滚都不能导致管理入口消失。
4. 所有发布和回滚操作都记录 ConfigVersion。

---

# 十一、前端实现要求

前端使用 frontend 模板原有架构。

技术要求：

1. React
2. Vite
3. TypeScript
4. Tailwind CSS
5. shadcn/ui
6. React Router
7. Zustand
8. React Hook Form
9. Zod

---

# 十二、前端 API

新增：

```text
src/api/client.ts
src/api/proxy-sites.ts
src/api/caddy.ts
src/api/config-versions.ts
src/api/dashboard.ts
```

要求：

1. API_BASE_URL 默认使用空字符串。
2. 因为生产环境前端和后端同域，API 请求走 /api。
3. 开发环境可以通过 VITE_API_BASE_URL 覆盖。
4. 请求自动带 JWT。
5. 401 自动跳转登录页。

---

# 十三、前端页面

路由：

```text
/login
/dashboard
/proxy-sites
/proxy-sites/new
/proxy-sites/:id/edit
/proxy-sites/:id/clone
/config-versions
/config-versions/:id
/caddy
/settings
```

左侧菜单：

```text
仪表盘
代理站点
配置版本
Caddy 状态
系统设置
```

---

# 十四、代理站点页面

列表字段：

```text
名称
域名
上游
HTTPS
状态
更新时间
操作
```

操作：

```text
编辑
克隆
启用 / 停用
预览配置
删除
```

---

# 十五、代理站点表单

使用 React Hook Form + Zod。

字段：

```text
名称
描述
启用状态
域名，多行输入，每行一个
上游，多行输入，每行一个
启用 HTTPS
强制 HTTPS
启用 WebSocket
启用 gzip / zstd
启用访问日志
请求头 JSON
响应头 JSON
IP 白名单，多行输入
Basic Auth 开关
Basic Auth 用户 JSON
advanced_json 文本框
```

按钮：

```text
保存
保存并发布
预览 Caddy JSON
取消
```

---

# 十六、配置版本页面

列表字段：

```text
版本号
状态
原因
发布时间
创建时间
操作
```

操作：

```text
查看 JSON
回滚
```

详情页展示：

```text
business_config
caddy_json
error_message
```

---

# 十七、Caddy 状态页面

展示：

```text
Admin API 地址
连接状态
当前 Caddy JSON
刷新按钮
发布按钮
```

---

# 十八、Dashboard 页面

展示：

```text
站点总数
启用站点
停用站点
HTTPS 站点
Caddy 状态
最近发布时间
```

---

# 十九、Dockerfile

根目录创建 Dockerfile。

必须是多阶段构建。

阶段 1：构建前端。

```text
node 镜像
进入 frontend
pnpm install
pnpm build
输出 dist
```

阶段 2：构建后端。

```text
golang 镜像
进入 backend
go build
输出 caddypilot
```

阶段 3：最终运行镜像。

最终镜像必须包含：

```text
caddy
supervisor
ca-certificates
tzdata
/app/backend/caddypilot
/app/frontend
/etc/caddy/Caddyfile
/etc/supervisor/conf.d/supervisord.conf
/data
```

---

# 二十、docker/Caddyfile

创建：

```text
docker/Caddyfile
```

内容：

```text
{
    admin 127.0.0.1:2019
    persist_config off
    auto_https disable_redirects
}

:8080 {
    root * /app/frontend
    encode gzip zstd

    handle_path /api/* {
        reverse_proxy 127.0.0.1:25610
    }

    handle {
        try_files {path} /index.html
        file_server
    }
}
```

---

# 二十一、docker/supervisord.conf

创建：

```text
docker/supervisord.conf
```

要求：

1. 启动 caddy。
2. 启动 caddypilot backend。
3. 日志输出到 stdout / stderr。
4. 不使用 daemon 模式。

---

# 二十二、docker/entrypoint.sh

创建：

```text
docker/entrypoint.sh
```

要求：

1. 创建 /data。
2. 创建 /data/caddy。
3. 设置默认环境变量。
4. 启动 supervisord。

---

# 二十三、docker-compose.yml

根目录创建：

```yaml
services:
  caddypilot:
    image: caddypilot:latest
    build:
      context: .
      dockerfile: Dockerfile
    container_name: caddypilot
    restart: unless-stopped
    ports:
      - "8080:8080"
      - "80:80"
      - "443:443"
    environment:
      TZ: Asia/Shanghai
      DATABASE_DSN: /data/caddypilot.db
      CADDY_ADMIN_API: http://127.0.0.1:2019
      CADDYPILOT_FRONTEND_DIR: /app/frontend
      CADDYPILOT_BACKEND_ADDR: 127.0.0.1:25610
      CADDYPILOT_MANAGE_ADDR: :8080
    volumes:
      - ./data:/data
```

要求：

1. 不要映射 2019 端口。
2. 生产环境说明不要暴露 2019。
3. 默认通过 http://localhost:8080 访问管理界面。

---

# 二十四、文档

创建：

```text
README.md
docs/design.md
docs/api.md
docs/caddy-json.md
docs/deployment.md
```

文档必须使用中文。

README 必须包含：

1. 项目介绍
2. 功能范围
3. 不做什么
4. 架构说明
5. 单镜像部署
6. docker compose 使用方法
7. 默认访问地址
8. Caddy Admin API 安全提醒
9. 数据目录说明
10. 开发环境启动
11. 后续计划

必须明确说明：

```text
不要将 Caddy Admin API 2019 端口暴露到公网。
```

---

# 二十五、质量要求

后端必须通过：

```bash
cd backend
go test ./...
```

前端必须通过：

```bash
cd frontend
pnpm install
pnpm build
```

整体必须可以运行：

```bash
docker compose up -d --build
```

访问：

```text
http://localhost:8080
```

验收：

1. 可以登录。
2. 可以新增代理站点。
3. 可以编辑代理站点。
4. 可以克隆代理站点。
5. 克隆后的站点默认 disabled。
6. 可以启用 / 停用代理站点。
7. 可以预览 Caddy JSON。
8. 可以发布 Caddy JSON 到 Caddy。
9. 发布后管理界面不能失联。
10. 可以查看配置版本。
11. 可以回滚配置版本。
12. 回滚后管理界面不能失联。
13. 2019 端口没有暴露到宿主机。
14. Docker 镜像内包含前端、后端和 Caddy。

---

# 二十六、提交要求

每完成一个阶段，进行一次 git commit。

推荐 commit：

```text
init project from starter templates
add project plan
add backend models
add proxy site api
add caddy json generator
add caddy admin client
add config publish and rollback
add frontend api client and layout
add proxy site pages
add config version pages
add caddy status dashboard
add single image docker deployment
add documentation
fix tests and build
```

---

# 二十七、执行方式

请你自动完成以上所有步骤。

不要只输出说明。

需要直接修改项目文件、创建代码、创建文档、执行测试和构建。

如果遇到模板结构和预期不一致，以实际模板结构为准，但必须保持功能目标不变。

如果某个高级功能实现成本过高，可以先做最小可用实现，但必须在 README 和 plan.md 中标记为 TODO。

最终必须保证：

```bash
go test ./...
pnpm build
docker compose up -d --build
```

尽可能可用。
