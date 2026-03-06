#!/bin/bash
# Nginx 安装与配置脚本（服务器上执行一次）
# 使用方法：sudo bash setup-nginx.sh

set -e

DOMAIN="web3-test.dreambinjk.top"
NGINX_CONF="/etc/nginx/sites-available/fss-mining"

echo "========================================"
echo "  FSS 挖矿系统 - Nginx 配置"
echo "========================================"

# 1. 安装 Nginx
if ! command -v nginx &>/dev/null; then
    apt update
    apt install -y nginx
    echo "✅ Nginx 安装完成"
else
    echo "✅ Nginx 已安装"
fi

# 2. 写入配置文件
cat > $NGINX_CONF << EOF
server {
    listen 80;
    server_name $DOMAIN;

    access_log /var/log/nginx/fss-mining-access.log;
    error_log  /var/log/nginx/fss-mining-error.log;

    client_max_body_size 10m;

    location / {
        proxy_pass         http://127.0.0.1:8888;
        proxy_http_version 1.1;

        proxy_set_header   Host              \$host;
        proxy_set_header   X-Real-IP         \$remote_addr;
        proxy_set_header   X-Forwarded-For   \$proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto \$scheme;

        proxy_connect_timeout  10s;
        proxy_read_timeout     60s;
        proxy_send_timeout     60s;
    }
}
EOF

echo "✅ Nginx 配置已写入: $NGINX_CONF"

# 3. 启用站点（创建软链接）
ln -sf $NGINX_CONF /etc/nginx/sites-enabled/fss-mining

# 4. 删除默认站点（避免端口冲突）
rm -f /etc/nginx/sites-enabled/default

# 5. 测试配置语法
nginx -t

# 6. 重载 Nginx
systemctl reload nginx
systemctl enable nginx

echo ""
echo "========================================"
echo "✅ Nginx 配置完成！"
echo ""
echo "  访问地址："
echo "  API 文档：http://$DOMAIN/swagger/index.html"
echo "  健康检查：http://$DOMAIN/health"
echo "========================================"
