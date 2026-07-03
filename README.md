# lmvpn_server

## 目录说明

- `data/` - 程序运行生成的数据（配置文件、数据库文件）
- `dist/` - 前端构建产物，由 Go 服务端托管提供静态资源
- `frontend/` - 前端工程（Vue 3 + TypeScript + Vite）
- `internal/` - Go 后端源码
  - `config/` - 配置加载
  - `db/` - 数据库操作
  - `handler/` - HTTP 请求处理
  - `middleware/` - 中间件（认证等）
  - `model/` - 数据模型
  - `vpn/` - VPN 相关逻辑（认证、隧道等）
- `pytest/` - Python 测试脚本
- `main.go` - Go 服务端入口文件
- `go.mod` / `go.sum` - Go 模块依赖管理

## 安全配置

### JWT 密钥

JWT 密钥按以下优先级加载：

1. 环境变量 `LMVPN_JWT_SECRET`
2. 配置文件 `data/config.yml` 中的 `web.jwt_secret`
3. 首次启动时自动生成 32 字节随机密钥并写入配置文件

生产环境建议通过环境变量注入，避免密钥落盘。

### 默认管理员

首次启动时自动创建管理员账户，密码为随机生成的 16 位字符串。密码会：

- 打印到 stdout（仅一次）
- 写入 `data/.initial_admin_password`（权限 0600）

请登录后立即修改密码，并删除 `data/.initial_admin_password` 文件。

### 登录限流

`/api/login` 和 WebSocket 密码认证均限制每 IP 5 次/分钟。

### Unix Socket 权限

默认权限兼容反向代理（如 Caddy），可通过配置文件调整：

```yaml
web:
  sock: "/run/lmvpnweb.sock"
  sock_mode: "0666"        # socket 文件权限，默认 0666
  sock_group: ""           # socket 文件 group，空=不修改
  sock_dir_mode: "0755"    # socket 目录权限，默认 0755
```

多租户/高安全场景建议收紧：

```yaml
web:
  sock: "/run/lmvpnweb.sock"
  sock_mode: "0660"
  sock_group: "caddy"      # 将 lmvpn 进程用户加入 caddy group
  sock_dir_mode: "0750"
```

### 生产部署

- 生产环境必须通过反向代理（如 Caddy/Nginx）提供 HTTPS/WSS
- 配置文件 `data/config.yml` 权限为 0600，仅限运行用户读写
- WebSocket `/ws` 端点支持 JWT 认证（`?token=xxx`）和密码认证两种方式
