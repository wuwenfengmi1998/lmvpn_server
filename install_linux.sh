#!/usr/bin/env bash
set -e

# ─────────────────────────────────────────────────────────────
# VPN 子网（与后台 VPN 设置中的子网保持一致）
# 若后台修改了子网，需同步修改此处并重新执行脚本，或手动更新 iptables 规则
# ─────────────────────────────────────────────────────────────
VPN_SUBNET="192.168.77.0/24"

if [ "$(id -u)" -ne 0 ]; then
    echo "请使用 root 用户执行此脚本: sudo bash install_linux.sh"
    exit 1
fi

echo ">>> 拉取最新代码..."
git pull

echo ">>> 安装前端依赖并构建..."
cd frontend
npm install
npm run build
cd ..

echo ">>> 编译 Go 后端..."
go build -o lmvpn .

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
# 持久化
cat > /etc/sysctl.d/99-lmvpn.conf << EOF
net.ipv4.ip_forward = 1
EOF
sysctl -p /etc/sysctl.d/99-lmvpn.conf >/dev/null

echo ">>> 配置 iptables NAT 与转发规则..."
# 自动检测默认路由出口网卡
WAN_IFACE=$(ip route show default 2>/dev/null | awk '{print $5; exit}')
if [ -z "$WAN_IFACE" ]; then
    echo "警告: 未能自动检测出口网卡，跳过 iptables 配置，请手动设置"
else
    echo "出口网卡: $WAN_IFACE"

    # 幂等：先删除旧规则（忽略错误），再添加
    # NAT: VPN 子网经出口网卡做源地址转换
    iptables -t nat -D POSTROUTING -s "$VPN_SUBNET" -o "$WAN_IFACE" -j MASQUERADE 2>/dev/null || true
    iptables -t nat -A POSTROUTING -s "$VPN_SUBNET" -o "$WAN_IFACE" -j MASQUERADE

    # FORWARD: 放行 VPN 子网进出
    iptables -D FORWARD -s "$VPN_SUBNET" -j ACCEPT 2>/dev/null || true
    iptables -D FORWARD -d "$VPN_SUBNET" -j ACCEPT 2>/dev/null || true
    iptables -A FORWARD -s "$VPN_SUBNET" -j ACCEPT
    iptables -A FORWARD -d "$VPN_SUBNET" -j ACCEPT

    # 持久化 iptables 规则
    if command -v netfilter-persistent >/dev/null 2>&1; then
        netfilter-persistent save
        echo "iptables 规则已通过 netfilter-persistent 持久化"
    elif command -v iptables-save >/dev/null 2>&1; then
        mkdir -p /etc/iptables
        iptables-save > /etc/iptables/rules.v4
        echo "iptables 规则已写入 /etc/iptables/rules.v4"
        echo "注意: 重启后规则持久化需配合 iptables-persistent 包，建议安装: apt install iptables-persistent"
    else
        echo "警告: 未找到 iptables-save，规则仅在本次运行有效，请手动持久化"
    fi
fi

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
systemctl status lmvpn --no-pager
