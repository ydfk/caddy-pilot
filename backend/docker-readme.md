# Docker 使用说明

## 开发模式

Win11 推荐使用开发镜像，容器内已安装 GCC 和 Air：

```bash
docker compose -f docker-compose.dev.yml up --build
```

源码挂载到 `/app`，Go 模块与构建缓存使用命名卷。修改 Go 或 YAML 文件后会自动重新编译。

```bash
docker compose -f docker-compose.dev.yml down
```

## 生产模式

```bash
docker compose up -d --build
docker compose logs -f
```

默认端口为 `25610`：

- API 文档：`http://localhost:25610/docs`
- OpenAPI 3.1：`http://localhost:25610/openapi.yaml`

生产镜像使用多阶段构建。构建阶段启用 CGO 编译官方 SQLite 驱动，运行阶段只保留应用、配置、CA 证书和时区数据，并使用非 root 用户。

## 数据持久化

- `app-data`：SQLite 数据
- `app-logs`：应用日志
- `./config:/app/config`：运行配置

停止服务时不要添加 `-v`，除非确认需要删除持久化数据。

## 构建镜像源

默认值适合国内网络，也可以按环境覆盖：

```bash
docker build \
  --build-arg GOPROXY=https://proxy.golang.org,direct \
  --build-arg APK_MIRROR=dl-cdn.alpinelinux.org \
  -t go-fiber-starter .
```
