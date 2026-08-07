# Passkey、Caddy 更新网络与站点跳转实施计划

## 目标

1. 保留用户名密码登录，并增加 Passkey 注册、列表、删除和无密码登录。
2. Passkey 的管理 Origin、RP ID、RP 名称和允许 Origin 均由配置文件或环境变量提供，不在认证逻辑中写死部署地址。
3. 修复 Docker 内 `127.0.0.11` DNS 异常导致 GitHub 更新检查失败的问题，Compose 默认使用 `223.5.5.5`，并允许部署者覆盖。
4. 代理站点列表中的有效域名可直接在新标签页打开；HTTP/HTTPS 外部端口从环境变量读取，80/443 不附加端口。

## 实施步骤

1. 扩展后端配置：加入公开管理 Origin、Passkey RP ID、RP 名称和允许 Origin；支持 YAML 与环境变量覆盖，并校验 Origin/RP ID。
2. 引入 WebAuthn 服务：封装注册与登录挑战、短期会话、凭据转换和签名计数更新；挑战只在短期内有效且一次性消费。
3. 新增 Passkey 凭据模型及自动迁移，凭据归属现有单管理员用户；凭据公开响应不包含公钥等敏感字段。
4. 扩展认证 API：
   - 登录后开始/完成 Passkey 注册；
   - 登录后列出、重命名和删除 Passkey；
   - 未登录开始/完成 Passkey 登录并签发原有 JWT。
5. 扩展登录页：保留密码表单，已有 Passkey 时显示无密码登录入口；处理浏览器不支持、安全上下文不满足和用户取消等错误。
6. 新增账户安全页并加入侧栏：显示当前 Passkey、添加、重命名和删除，明确密码仍可作为恢复方式。
7. 扩展系统信息 API，返回公开 HTTP/HTTPS 端口；Compose 同时把 `CADDYPILOT_HTTP_PORT` 和 `CADDYPILOT_HTTPS_PORT` 传入容器。
8. 修改代理站点域名列：为非通配符域名生成目标 URL；HTTPS 站点使用 HTTPS 端口，其余使用 HTTP 端口；标准端口省略，非标准端口显式附加。
9. 修改两个 Compose 文件：默认 DNS 为 `223.5.5.5`，通过 `CADDYPILOT_DNS_SERVER` 覆盖；补齐 Passkey 与公开端口环境变量透传。
10. 更新 README 与部署文档，说明 DNS、Passkey HTTPS 前提、配置示例和端口跳转行为。
11. 增加后端单元/API 测试和前端组件/工具测试，运行 Go 测试、前端测试、lint、格式检查和构建；记录任何仅能在真实浏览器验证的边界。

## 验收标准

- 原密码注册、登录、JWT 保护接口保持可用。
- 已登录管理员可添加多个 Passkey，并能列出、重命名、删除。
- 登录页能使用已登记 Passkey 获得与密码登录相同的 JWT。
- Passkey 配置缺失或 Origin 不合法时返回明确错误，不降低为不校验 Origin。
- Compose 默认绕过异常的 Docker 内置 DNS 上游，且可通过环境变量替换 DNS。
- `CADDYPILOT_HTTP_PORT=18080`、`CADDYPILOT_HTTPS_PORT=18443` 时，域名分别链接到 `http://域名:18080` 与 `https://域名:18443`；80/443 时不显示端口。
- 通配符和自定义模式中没有明确域名的内容不生成无效链接。
- 后端测试、前端测试、lint、格式检查和构建通过。
