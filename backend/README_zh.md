# Go Fiber Starter

基于 Fiber v3、Huma 和 GORM 的 Go API 启动模板。项目采用 OpenAPI 3.1 code-first：先编写 Go 类型和处理函数，运行应用后自动提供 OpenAPI 文档，不需要维护规范文件，也不会从规范反向生成业务代码。

[English](README.md)

## 主要特性

- Fiber v3 HTTP 服务
- Huma v2 code-first API 与 OpenAPI 3.1
- 运行时自动提供 API 文档和 JSON/YAML 规范
- JWT Bearer 认证
- GORM 官方 SQLite、PostgreSQL、MySQL 驱动
- SQLite 使用官方 `gorm.io/driver/sqlite`
- 分层 YAML 配置
- Zap 日志与数据库自动迁移
- Win11 原生开发和 Docker 热更新开发
- Linux Alpine 多阶段生产镜像

## OpenAPI 3.1

启动应用后可直接访问：

- 文档界面：`http://localhost:25610/docs`
- OpenAPI 3.1 JSON：`http://localhost:25610/openapi.json`
- OpenAPI 3.1 YAML：`http://localhost:25610/openapi.yaml`
- OpenAPI 3.0.3 兼容输出：`/openapi-3.0.json`、`/openapi-3.0.yaml`

项目没有 `docs/` 生成目录，也不需要运行 `swag init`。API 结构来自 Go 代码中的输入、输出类型及 `huma.Operation`：

```go
huma.Register(api, huma.Operation{
	OperationID: "get-user-profile",
	Method:      http.MethodGet,
	Path:        "/api/auth/profile",
	Summary:     "获取当前用户",
	Security: []map[string][]string{
		{BearerAuthScheme: {}},
	},
}, Profile)
```

新增接口时，定义强类型输入/输出并注册处理函数即可。应用启动后，文档与规范会立即反映代码变化。

## 快速开始

### Docker 开发，推荐用于 Win11

只需要 Docker Desktop，不要求 Windows 安装 GCC、Go 或 Air：

```bat
scripts\dev-docker.bat
```

等价命令：

```bash
docker compose -f docker-compose.dev.yml up --build
```

源码会挂载到容器，Air 自动重新编译和启动服务。停止服务：

```bash
docker compose -f docker-compose.dev.yml down
```

### Win11 原生开发

官方 SQLite 驱动使用 `go-sqlite3`，因此必须启用 CGO，并安装 MinGW-w64 或其他可用的 GCC。确保 `gcc.exe` 已加入 `PATH`：

```powershell
go env -w CGO_ENABLED=1
gcc --version
```

安装 Air 后启动热更新：

```powershell
go install github.com/air-verse/air@v1.65.3
scripts\dev.bat
```

其他脚本：

```bat
scripts\build.bat
scripts\run.bat
scripts\test.bat
```

脚本会先检查 GCC；没有本机编译器时请使用 Docker 开发模式。

### 生产运行

```bash
docker compose up -d --build
docker compose logs -f
```

生产镜像在 Linux Alpine 构建阶段启用 CGO，运行阶段使用非 root 用户。

## API

| 方法 | 路径 | 说明 | 认证 |
| --- | --- | --- | --- |
| `POST` | `/api/auth/register` | 注册用户 | 否 |
| `POST` | `/api/auth/login` | 登录并获取 JWT | 否 |
| `GET` | `/api/auth/profile` | 获取当前用户 | Bearer JWT |

Huma 会验证请求体，并返回真实 HTTP 状态码，例如 `201`、`401`、`409` 和 `422`。

## 配置

基础配置位于 `config/config.yaml`。加载顺序如下，后加载的文件覆盖前面的值：

1. `config/config.yaml`
2. `config/config.<env>.yaml`
3. `config/config.local.yaml`
4. `config/config.<env>.local.yaml`

本机私有配置可参考 `config/config.local.yaml.example`。

```yaml
app:
  port: "25610"
  env: "development"
jwt:
  secret: "replace-in-production"
  expiration: 604800
database:
  driver: "sqlite"
  path: "data/db.sqlite"
  dsn: ""
```

数据库 `driver` 支持 `sqlite`、`postgres`、`postgresql` 和 `mysql`。

## 项目结构

```text
go-fiber-starter/
├── cmd/                         # 应用入口与 Huma 初始化
├── config/                      # YAML 配置
├── internal/
│   ├── api/auth/
│   │   ├── handler.go           # 强类型处理函数
│   │   ├── middleware.go        # Huma JWT 中间件
│   │   ├── router.go            # code-first Operation 注册
│   │   └── types.go             # OpenAPI 输入/输出模型
│   ├── model/                   # 数据模型
│   └── service/                 # 业务服务
├── pkg/                         # 配置、数据库、日志和工具
├── scripts/                     # Win11 开发脚本
├── Dockerfile                   # Linux 生产镜像
├── Dockerfile.dev               # CGO + Air 开发镜像
├── docker-compose.yml           # 生产运行
└── docker-compose.dev.yml       # 开发热更新
```

## 测试

Win11 原生环境需要 GCC：

```bash
go test ./...
go vet ./...
```

不安装 GCC 时可直接在开发镜像中验证：

```powershell
docker run --rm -e CGO_ENABLED=1 `
  -v "${PWD}:/app" -w /app go-fiber-starter:dev `
  sh -c "go test ./... && go vet ./..."
```

测试覆盖认证流程、JWT 校验、数据库配置，以及自动生成的 OpenAPI 3.1 规范和文档路由。

## 依赖选择

- `github.com/danielgtaylor/huma/v2`：直接适配 Fiber v3，运行时生成 OpenAPI 3.1。
- `gorm.io/driver/sqlite`：GORM 官方维护的 SQLite 驱动；需要 CGO。
- `gorm.io/driver/postgres`、`gorm.io/driver/mysql`：保留官方数据库驱动。
- 不再使用 Swaggo、Swagger 2 生成文件、Fiber contrib JWT 和 `github.com/glebarez/sqlite`。

Docker 构建参数 `GOPROXY` 和 `APK_MIRROR` 均可覆盖：

```bash
docker build \
  --build-arg GOPROXY=https://proxy.golang.org,direct \
  --build-arg APK_MIRROR=dl-cdn.alpinelinux.org \
  -t go-fiber-starter .
```

## License

[MIT](LICENSE)
