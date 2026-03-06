#!/bin/bash
# 服务器首次初始化脚本（仅首次部署时执行一次）
# 使用方法：sudo bash setup-server.sh

set -e

APP_USER="fss"
DEPLOY_DIR="/opt/fss-mining"
SERVICE_NAME="fss-mining"

echo "========================================"
echo "  FSS 挖矿系统 - 服务器初始化"
echo "========================================"

# 1. 创建专用运行用户
if ! id "$APP_USER" &>/dev/null; then
    useradd -r -s /bin/false -d $DEPLOY_DIR "$APP_USER"
    echo "✅ 创建用户: $APP_USER"
fi

# 2. 创建部署目录
mkdir -p $DEPLOY_DIR/configs
mkdir -p $DEPLOY_DIR/logs
chown -R $APP_USER:$APP_USER $DEPLOY_DIR
echo "✅ 创建目录: $DEPLOY_DIR"

# 3. 复制配置文件（首次需要手动上传并修改 configs/config.yaml）
if [ ! -f "$DEPLOY_DIR/configs/config.yaml" ]; then
    echo "⚠️  请手动上传配置文件到 $DEPLOY_DIR/configs/config.yaml"
fi

# 4. 安装 systemd 服务
cp "$(dirname "$0")/fss-mining.service" /etc/systemd/system/$SERVICE_NAME.service
systemctl daemon-reload
systemctl enable $SERVICE_NAME
echo "✅ systemd 服务已注册: $SERVICE_NAME"

# 5. 配置 sudo 权限（让 deploy 用户可以重启服务，无需密码）
DEPLOY_USER="${SUDO_USER:-ubuntu}"
SUDOERS_FILE="/etc/sudoers.d/fss-deploy"
echo "$DEPLOY_USER ALL=(ALL) NOPASSWD: /bin/systemctl restart $SERVICE_NAME, /bin/systemctl status $SERVICE_NAME, /bin/systemctl is-active $SERVICE_NAME" > $SUDOERS_FILE
chmod 0440 $SUDOERS_FILE
echo "✅ sudo 权限已配置（用户: $DEPLOY_USER）"

echo ""
echo "========================================"
echo "  初始化完成！"
echo ""
echo "  后续操作："
echo "  1. 上传并修改配置文件："
echo "     $DEPLOY_DIR/configs/config.yaml"
echo ""
echo "  2. 在 GitHub 仓库 Settings > Secrets 中添加："
echo "     SSH_PRIVATE_KEY  - 服务器 SSH 私钥"
echo "     SERVER_HOST      - 服务器 IP 或域名"
echo "     SSH_USER         - SSH 登录用户名（如 ubuntu）"
echo ""
echo "  3. 推送到 main 分支即可自动部署"
echo "========================================"
