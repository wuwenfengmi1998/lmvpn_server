#!/usr/bin/env bash
set -e

if [ "$(id -u)" -ne 0 ]; then
    echo "请使用 root 用户执行此脚本: sudo bash install_linux.sh"
    exit 1
fi

echo ">>> 拉取最新代码（强制覆盖本地修改）..."
git fetch origin
git reset --hard origin/main

echo ">>> 安装前端依赖并构建..."
cd frontend
npm install
npm run build
cd ..

echo ">>> 编译 Go 后端..."
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD)
COMMIT_TIME=$(git log -1 --format=%cI)
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
go build -ldflags "-X lmvpn/internal/version.Version=$VERSION -X lmvpn/internal/version.Commit=$COMMIT -X lmvpn/internal/version.CommitTime=$COMMIT_TIME -X lmvpn/internal/version.BuildTime=$BUILD_TIME" -o lmvpn .

echo ">>> 创建 lmvpn 系统用户..."
if ! id lmvpn &>/dev/null; then
    useradd -r -s /bin/false lmvpn
    echo "用户 lmvpn 已创建"
else
    echo "用户 lmvpn 已存在，跳过创建"
fi

echo ">>> 部署到 /opt/lmvpn..."
systemctl stop lmvpn 2>/dev/null || true
mkdir -p /opt/lmvpn
cp lmvpn /opt/lmvpn/
cp -r dist /opt/lmvpn/
chown -R lmvpn:lmvpn /opt/lmvpn

echo ">>> 配置内核 IP 转发..."
# 临时生效
sysctl -w net.ipv4.ip_forward=1 >/dev/null
sysctl -w net.ipv6.conf.all.forwarding=1 >/dev/null
# 持久化
cat > /etc/sysctl.d/99-lmvpn.conf << EOF
net.ipv4.ip_forward = 1
net.ipv6.conf.all.forwarding = 1
EOF
sysctl -p /etc/sysctl.d/99-lmvpn.conf >/dev/null

echo ">>> 安装 systemd 服务..."
cat > /etc/systemd/system/lmvpn.service << 'EOF'
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

# 授予 TUN 设备创建与网络管理能力，无需 root 运行
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_RAW
CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_RAW

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable lmvpn
systemctl restart lmvpn

echo ">>> 安装完成"
echo ""
echo "NAT/转发/UFW 规则由服务端程序在启动时根据后台子网配置自动管理，无需手动设置。"
echo "若修改了后台 VPN 子网，只需在后台保存即可，程序会自动更新防火墙规则。"
echo ""
systemctl status lmvpn --no-pager
