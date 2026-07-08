#!/usr/bin/env bash
set -e

# ─────────────────────────────────────────────────────────────
# VPN 子网（与后台 VPN 设置中的子网保持一致）
# 若后台修改了子网，需同步修改此处并重新执行脚本，或手动更新 iptables 规则
# ─────────────────────────────────────────────────────────────
VPN_SUBNET="192.168.77.0/24"
# IPv6 子网（留空则不配置 IPv6 NAT；与后台设置保持一致）
VPN_SUBNET6="fd00:dead:beef::/112"

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
if [ -n "$VPN_SUBNET6" ]; then
    sysctl -w net.ipv6.conf.all.forwarding=1 >/dev/null
fi
# 持久化
cat > /etc/sysctl.d/99-lmvpn.conf << EOF
net.ipv4.ip_forward = 1
EOF
if [ -n "$VPN_SUBNET6" ]; then
    echo "net.ipv6.conf.all.forwarding = 1" >> /etc/sysctl.d/99-lmvpn.conf
fi
sysctl -p /etc/sysctl.d/99-lmvpn.conf >/dev/null

echo ">>> 配置 NAT 与转发规则..."
# 自动检测默认路由出口网卡
WAN_IFACE=$(ip route show default 2>/dev/null | awk '{print $5; exit}')
if [ -z "$WAN_IFACE" ]; then
    echo "警告: 未能自动检测出口网卡，跳过 NAT 配置，请手动设置"
else
    echo "出口网卡: $WAN_IFACE"

    # 选择可用的防火墙前端：优先 nft（原生，Debian 12+ 的 iptables 是 nft 包装器，操作原生 nft nat 表会不兼容）
    NAT_TOOL=""
    if command -v nft >/dev/null 2>&1; then
        NAT_TOOL="nft"
    elif command -v iptables >/dev/null 2>&1; then
        NAT_TOOL="iptables"
    fi

    if [ -z "$NAT_TOOL" ]; then
        echo "警告: 未找到 nft 或 iptables，跳过 NAT 配置"
        echo "       客户端将无法上外网。请安装: apt install nftables"
        echo "       安装后重新执行本脚本即可自动配置"
    elif [ "$NAT_TOOL" = "nft" ]; then
        echo "使用 nft 配置 NAT..."
        # 幂等：先删除旧表（忽略错误），再创建
        nft delete table inet lmvpn_nat 2>/dev/null || true
        nft add table inet lmvpn_nat
        nft 'add chain inet lmvpn_nat postrouting { type nat hook postrouting priority 100 ; }'
        nft add rule inet lmvpn_nat postrouting oifname "$WAN_IFACE" ip saddr "$VPN_SUBNET" masquerade
        if [ -n "$VPN_SUBNET6" ]; then
            nft add rule inet lmvpn_nat postrouting oifname "$WAN_IFACE" ip6 saddr "$VPN_SUBNET6" masquerade
        fi
        nft 'add chain inet lmvpn_nat forward { type filter hook forward priority 0 ; policy accept ; }'
        nft add rule inet lmvpn_nat forward ip saddr "$VPN_SUBNET" accept
        nft add rule inet lmvpn_nat forward ip daddr "$VPN_SUBNET" accept
        if [ -n "$VPN_SUBNET6" ]; then
            nft add rule inet lmvpn_nat forward ip6 saddr "$VPN_SUBNET6" accept
            nft add rule inet lmvpn_nat forward ip6 daddr "$VPN_SUBNET6" accept
        fi

        # 持久化 nft 规则
        mkdir -p /etc/nftables.d
        nft list ruleset > /etc/nftables.d/lmvpn.nft
        # 确保主配置文件 include 该目录
        if [ -f /etc/nftables.conf ] && ! grep -q 'include "/etc/nftables.d' /etc/nftables.conf; then
            echo 'include "/etc/nftables.d/*.nft"' >> /etc/nftables.conf
        fi
        systemctl enable nftables 2>/dev/null || true
        echo "nft 规则已写入 /etc/nftables.d/lmvpn.nft"
    elif [ "$NAT_TOOL" = "iptables" ]; then
        echo "使用 iptables 配置 NAT..."
        # 幂等：先删除旧规则（忽略错误），再添加
        iptables -t nat -D POSTROUTING -s "$VPN_SUBNET" -o "$WAN_IFACE" -j MASQUERADE 2>/dev/null || true
        iptables -t nat -A POSTROUTING -s "$VPN_SUBNET" -o "$WAN_IFACE" -j MASQUERADE

        iptables -D FORWARD -s "$VPN_SUBNET" -j ACCEPT 2>/dev/null || true
        iptables -D FORWARD -d "$VPN_SUBNET" -j ACCEPT 2>/dev/null || true
        iptables -A FORWARD -s "$VPN_SUBNET" -j ACCEPT
        iptables -A FORWARD -d "$VPN_SUBNET" -j ACCEPT

        # IPv6 NAT（需 ip6tables）
        if [ -n "$VPN_SUBNET6" ] && command -v ip6tables >/dev/null 2>&1; then
            echo "配置 IPv6 NAT (ip6tables)..."
            ip6tables -t nat -D POSTROUTING -s "$VPN_SUBNET6" -o "$WAN_IFACE" -j MASQUERADE 2>/dev/null || true
            ip6tables -t nat -A POSTROUTING -s "$VPN_SUBNET6" -o "$WAN_IFACE" -j MASQUERADE
            ip6tables -D FORWARD -s "$VPN_SUBNET6" -j ACCEPT 2>/dev/null || true
            ip6tables -D FORWARD -d "$VPN_SUBNET6" -j ACCEPT 2>/dev/null || true
            ip6tables -A FORWARD -s "$VPN_SUBNET6" -j ACCEPT
            ip6tables -A FORWARD -d "$VPN_SUBNET6" -j ACCEPT
        fi

        # 持久化 iptables 规则
        if command -v netfilter-persistent >/dev/null 2>&1; then
            netfilter-persistent save
            echo "iptables 规则已通过 netfilter-persistent 持久化"
        elif command -v iptables-save >/dev/null 2>&1; then
            mkdir -p /etc/iptables
            iptables-save > /etc/iptables/rules.v4
            if [ -n "$VPN_SUBNET6" ] && command -v ip6tables-save >/dev/null 2>&1; then
                ip6tables-save > /etc/iptables/rules.v6
                echo "ip6tables 规则已写入 /etc/iptables/rules.v6"
            fi
            echo "iptables 规则已写入 /etc/iptables/rules.v4"
            echo "注意: 重启后规则持久化需配合 iptables-persistent 包，建议安装: apt install iptables-persistent"
        else
            echo "警告: 未找到 iptables-save，规则仅在本次运行有效，请手动持久化"
        fi
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
