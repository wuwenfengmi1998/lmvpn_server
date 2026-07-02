#!/usr/bin/env bash
set -e

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
mkdir -p /opt/lmvpn
cp lmvpn /opt/lmvpn/
cp -r dist /opt/lmvpn/
chown -R lmvpn:lmvpn /opt/lmvpn

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

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable lmvpn

echo ">>> 安装完成"
echo "启动服务: systemctl start lmvpn"
echo "查看状态: systemctl status lmvpn"
