# LMVPN

基于 WebSocket 隧道与 TUN 虚拟网卡的轻量级三层 VPN 系统，采用 Go + Vue 3 构建。服务端通过 WebSocket 与客户端建立控制通道，TUN 网卡与 WebSocket 之间双向搬运原始 IP 数据包，实现点对点与点对站点网络连接。

> **平台说明**：服务端部署**目前仅在 Linux 下测试功能正常**（systemd + nftables/iptables NAT，自动兼容 UFW）。配套客户端见 [lmvpn_client](https://github.com/wuwenfengmi1998/lmvpn_client)，客户端协议规范见 [`docs/client-development.md`](docs/client-development.md)。

---

## 目录

- [环境要求](#环境要求)
- [一键部署（Linux，推荐）](#一键部署linux推荐)
- [首次登录与初始化](#首次登录与初始化)
- [手动部署](#手动部署)
- [配置说明](#配置说明)
- [反向代理（HTTPS/WSS）](#反向代理httpswss)
- [服务管理](#服务管理)
- [升级](#升级)
- [目录说明](#目录说明)
- [常见问题排查](#常见问题排查)
- [相关文档](#相关文档)
- [许可证](#许可证)

---

## 环境要求

- **操作系统**：Linux（已测试），需 root 或 sudo 权限
- **Go**：≥ 1.26.4
- **Node.js**：≥ 22.18（用于构建前端）
- **内核**：支持 TUN 设备（`/dev/net/tun`），开启 `ip_forward`
- **防火墙工具**：nftables（推荐）或 iptables，兼容 UFW
- **网络**：公网 IP，开放 Web 端口（或经反向代理）

---

## 一键部署（Linux，推荐）

仓库自带 `install_linux.sh`，一条命令完成构建、部署、内核转发配置与 systemd 服务安装。NAT / 转发 / UFW 规则由服务端程序在启动时根据后台子网配置自动管理，无需手动设置。

```bash
git clone <仓库地址> lmvpn_server
cd lmvpn_server
sudo bash install_linux.sh
```

### 脚本执行流程

1. 拉取最新代码（`git fetch origin && git reset --hard origin/main`）
2. 构建前端：`cd frontend && npm install && npm run build`（产物输出到 `../dist`）
3. 编译后端：`go build -o lmvpn .`
4. 创建无 shell 的系统用户 `lmvpn`
5. 停止旧服务，部署到 `/opt/lmvpn`（二进制 + `dist/`），属主改为 `lmvpn`
6. 开启内核 IP 转发（IPv4 / IPv6），写入 `/etc/sysctl.d/99-lmvpn.conf`
7. 安装 systemd 服务 `/etc/systemd/system/lmvpn.service`，以 `lmvpn` 用户运行并授予 `CAP_NET_ADMIN` / `CAP_NET_RAW`
8. 启动服务（NAT / 转发 / UFW 规则由程序在启动时根据后台子网自动配置）

> ⚠️ 脚本中的 `git reset --hard origin/main` 会**丢弃所有本地未提交改动**。部署前请确保工作区干净，或先提交/暂存。

### 防火墙规则自动管理

NAT masquerade、forward 放行、UFW 转发规则由服务端程序在 `ApplySettings()` 时根据当前后台 VPN 子网动态配置：

- 自动检测出口网卡（`ip route show default`）
- 配置 nft `lmvpn_nat` 表的 postrouting masquerade 和 forward accept 规则
- 检测 UFW 是否启用（存在 `ufw-user-forward` 链），若启用则自动创建 `lmvpn-fwd` / `lmvpn6-fwd` 链并注入 jump
- 在后台修改子网后保存即可，程序自动更新规则，无需重新执行脚本

---

## 首次登录与初始化

1. **获取初始密码**：首次启动时自动创建管理员 `admin`，密码为随机 16 位字符串，会
   - 打印到 stdout（仅一次）
   - 写入 `/opt/lmvpn/data/.initial_admin_password`（权限 0600）

   ```bash
   sudo cat /opt/lmvpn/data/.initial_admin_password
   ```

2. **登录**：浏览器访问 `http://<服务器IP>:8080`，使用 `admin` + 初始密码登录。

3. **立即修改密码**，并删除初始密码文件：

   ```bash
   sudo rm /opt/lmvpn/data/.initial_admin_password
   ```

4. **启用 VPN**：进入「管理后台 → VPN 管理」，确认子网（默认 `192.168.77.0/24` + IPv6 `fd00:dead:beef::/112`），打开「启用」并保存。

5. **诊断检查**：VPN 管理页的「系统环境检测」面板（对应 `GET /api/admin/vpn/diag`）会检测 ip_forward、NAT、UFW 转发规则、TUN 等，客户端无法上网时优先查看此处。

---

## 手动部署

如需自行控制部署细节，可参照以下步骤。

### 1. 构建产物

```bash
# 前端
cd frontend
npm install
npm run build        # 产物输出到 ../dist
cd ..

# 后端
go build -o lmvpn .
```

### 2. 部署文件

```bash
sudo useradd -r -s /bin/false lmvpn
sudo mkdir -p /opt/lmvpn
sudo cp lmvpn /opt/lmvpn/
sudo cp -r dist /opt/lmvpn/
sudo chown -R lmvpn:lmvpn /opt/lmvpn
```

### 3. 开启内核转发

```bash
sudo sysctl -w net.ipv4.ip_forward=1
sudo sysctl -w net.ipv6.conf.all.forwarding=1   # IPv6 双栈时需要

cat <<'EOF' | sudo tee /etc/sysctl.d/99-lmvpn.conf
net.ipv4.ip_forward = 1
net.ipv6.conf.all.forwarding = 1
EOF
sudo sysctl -p /etc/sysctl.d/99-lmvpn.conf
```

### 4. 安装 systemd 服务

> NAT / 转发 / UFW 规则由服务端程序在启动时自动配置，无需手动设置。以下命令仅在程序无法自动配置防火墙时作为参考。

<details>
<summary>手动配置 NAT（点击展开，通常不需要）</summary>

将 `WAN_IFACE` 替换为出口网卡，`VPN_SUBNET` 替换为 VPN 子网：

```bash
WAN_IFACE=eth0
VPN_SUBNET=192.168.77.0/24
VPN_SUBNET6=fd00:dead:beef::/112

sudo nft add table inet lmvpn_nat
sudo nft 'add chain inet lmvpn_nat postrouting { type nat hook postrouting priority 100 ; }'
sudo nft add rule inet lmvpn_nat postrouting oifname "$WAN_IFACE" ip saddr "$VPN_SUBNET" masquerade
sudo nft add rule inet lmvpn_nat postrouting oifname "$WAN_IFACE" ip6 saddr "$VPN_SUBNET6" masquerade
sudo nft 'add chain inet lmvpn_nat forward { type filter hook forward priority 0 ; policy accept ; }'
sudo nft add rule inet lmvpn_nat forward ip saddr "$VPN_SUBNET" accept
sudo nft add rule inet lmvpn_nat forward ip daddr "$VPN_SUBNET" accept
```

若启用了 UFW，还需放行 VPN 子网转发：

```bash
sudo ufw route allow from "$VPN_SUBNET"
sudo ufw route allow to "$VPN_SUBNET"
sudo ufw route allow from "$VPN_SUBNET6"
sudo ufw route allow to "$VPN_SUBNET6"
```

</details>

### 5. 安装 systemd 服务

```bash
sudo tee /etc/systemd/system/lmvpn.service >/dev/null <<'EOF'
[Unit]
Description=LMVPN Server
After=network.target

[Service]
Type=simple
User=lmvpn
WorkingDirectory=/opt/lmvpn
ExecStart=/opt/lmvpn/lmvpn
Restart=on-failure
RestartSec=5
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_RAW
CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_RAW

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now lmvpn
```

---

## 配置说明

配置文件位于 `data/config.yml`（相对工作目录，即 `/opt/lmvpn/data/config.yml`），首次启动时自动生成。

```yaml
web:
  port: 8080                      # TCP 监听端口，0 则不监听 TCP
  sock: "/run/lmvpnweb.sock"      # Unix Socket 路径，空则不监听 socket
  sock_mode: "0666"               # socket 文件权限
  sock_group: ""                  # socket 文件属组，空=不修改
  sock_dir_mode: "0755"           # socket 目录权限
  jwt_secret: ""                  # JWT 密钥，留空则按下面规则生成
database:
  type: sqlite                    # sqlite 或 mysql
  path: data/lmvpn.db             # sqlite 文件路径
  dsn: ""                         # mysql DSN（type=mysql 时必填）
```

> `web.port` 与 `web.sock` **至少配置一个**，两者都配则同时监听。生产环境建议仅留一个，由反向代理统一接入。

### JWT 密钥

按以下优先级加载：

1. 环境变量 `LMVPN_JWT_SECRET`
2. 配置文件 `web.jwt_secret`
3. 首次启动自动生成 32 字节随机密钥并写入配置文件

生产环境建议通过环境变量注入，避免密钥落盘：

```bash
# /etc/systemd/system/lmvpn.service 的 [Service] 段追加
Environment=LMVPN_JWT_SECRET=<你的密钥>
```

### Unix Socket 权限收紧

多租户/高安全场景建议收紧 socket 权限（将 lmvpn 进程用户加入反代用户组）：

```yaml
web:
  sock: "/run/lmvpnweb.sock"
  sock_mode: "0660"
  sock_group: "caddy"
  sock_dir_mode: "0750"
```

> 若服务以非 root 用户运行且 socket 目录不可写，可将 `sock` 指向用户可写目录（如 `/opt/lmvpn/run/lmvpnweb.sock`），或设 `sock: ""` 仅用 TCP。

---

## 反向代理（HTTPS/WSS）

生产环境**必须**通过反向代理提供 HTTPS/WSS。服务端默认监听 HTTP `:8080`，由反代终结 TLS。

### Caddy（自动 HTTPS）

```caddyfile
lmvpn.example.com {
    reverse_proxy 127.0.0.1:8080
}
```

Caddy 的 `reverse_proxy` 自动处理 WebSocket 升级，无需额外配置。

若使用 Unix Socket 接入：

```caddyfile
lmvpn.example.com {
    reverse_proxy unix//run/lmvpnweb.sock
}
```

### Nginx

```nginx
server {
    listen 443 ssl http2;
    server_name lmvpn.example.com;

    ssl_certificate     /etc/ssl/certs/lmvpn.pem;
    ssl_certificate_key /etc/ssl/private/lmvpn.key;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade           $http_upgrade;
        proxy_set_header Connection        "upgrade";
        proxy_set_header Host              $host;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
}
```

> 部署反代后，客户端应使用 `wss://lmvpn.example.com/ws` 连接。`/ws` 的长连接超时建议调大（如 3600s），避免空闲断连。

---

## 服务管理

```bash
sudo systemctl start lmvpn       # 启动
sudo systemctl stop lmvpn        # 停止
sudo systemctl restart lmvpn     # 重启
sudo systemctl status lmvpn      # 状态
sudo systemctl enable lmvpn      # 开机自启
```

查看实时日志：

```bash
sudo journalctl -u lmvpn -f
```

---

## 升级

最简方式：重新执行一键脚本，它会自动拉取最新代码、重新构建并重启服务。

```bash
cd lmvpn_server
git pull
sudo bash install_linux.sh
```

> 脚本会 `git reset --hard origin/main`，本地改动会被覆盖。`/opt/lmvpn/data/` 下的配置与数据库不受影响。

---

## 目录说明

- `data/` — 运行时数据（`config.yml`、`lmvpn.db`、初始密码文件），gitignore
- `dist/` — 前端构建产物，由 Go 服务端托管，gitignore
- `frontend/` — 前端工程（Vue 3 + TypeScript + Vite）
- `internal/` — Go 后端源码
  - `config/` — 配置加载
  - `db/` — 数据库初始化
  - `handler/` — HTTP 请求处理
  - `middleware/` — 中间件（认证、限流）
  - `model/` — 数据模型
  - `router/` — 路由
  - `vpn/` - VPN 核心（认证、隧道、TUN、包转发、防火墙自动配置）
- `docs/` — 文档
- `pytest/` — Python 测试脚本
- `install_linux.sh` — Linux 一键部署脚本
- `main.go` — 服务端入口

---

## 常见问题排查

| 现象 | 可能原因 | 排查建议 |
|------|----------|----------|
| 服务启动失败 | socket 目录不可写 / 端口占用 | `journalctl -u lmvpn`；设 `sock: ""` 或换可写目录 |
| 客户端连上但无法上网 | 未开 ip_forward / NAT 未生效 / UFW 拦截 | 管理后台诊断面板查看 UFW 状态和 NAT 检测结果 |
| 客户端连上但下行极慢（几十 bps） | UFW FORWARD 默认 DROP 拦截 TCP 包 | 诊断面板检查 UFW 转发规则；重启服务使程序自动配置 UFW |
| 登录提示「请求过于频繁」 | 触发限流（`/api/login` 5 次/分钟·IP） | 等待 1 分钟后重试 |
| 修改子网后客户端异常 | 旧防火墙规则残留 | 后台重新保存设置，程序自动更新 NAT / UFW 规则 |
| 忘记管理员密码 | — | 删除 `data/lmvpn.db` 重新初始化（会丢失所有数据），或直接改库 |

---

## 相关文档

- [LMVPN 客户端项目](https://github.com/wuwenfengmi1998/lmvpn_client) — 配套客户端源码
- [客户端开发协议规范](docs/client-development.md) — WebSocket 隧道协议、认证、握手、数据面、TUN 配置等完整规范

---

## 安全提示

- 配置文件 `data/config.yml` 权限为 0600，仅限运行用户读写
- `/api/login` 与 WebSocket 密码认证均限制每 IP 5 次/分钟
- WebSocket `/ws` 支持 JWT（`?token=xxx`）与用户名/密码两种认证
- 生产环境务必经反向代理启用 HTTPS/WSS

---

## 许可证

本项目基于 [MIT License](LICENSE) 开源。

Copyright (c) 2026 wuwenfengmi1998
