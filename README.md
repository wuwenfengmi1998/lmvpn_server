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
