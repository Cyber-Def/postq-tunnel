# 安装与配置指南 (PostQ-Tunnel)

## 0. 基础设施要求
- 已安装 **Go 1.24** 或更高版本（推荐使用 Go 1.26）。
- 具有独立公共IP地址的 **云端 VPS 服务器**。
- 指向 VPS IP 的 **域名** (例如 `yourdomain.com`)。
  - *重要：* 必须配置泛域名解析 (Wildcard record) `*.yourdomain.com`，以便客户端动态挂载子域名前缀 (例如 `api.yourdomain.com`)。

## 1. 编译二进制文件
本项目采用纯 Go 的*零依赖*架构：

```bash
# 编译公共边缘节点 (部署于 VPS)
go build -o qtunnel ./cmd/server/main.go

# 编译本地客户端代理 (运行于本地 PC/Mac)
go build -o qtun ./cmd/qtun/main.go
```

## 2. 配置 VPS (边缘网关)
将编译产生的 `qtunnel` 文件部署至云服务器：
```bash
export QTUN_DOMAIN="tunnels.yourdomain.com"
export QTUN_EMAIL="admin@yourdomain.com" # 用于自动申请 HTTPS 证书

# 启动服务器 (绑定 80 和 443 需要 root 权限)
sudo -E ./qtunnel
```
**必须开放的防火墙端口:**
- `80` & `443` — 用户浏览器请求入口。
- `4443` — 后量子加密 PQC TLS 通信端口 (专供代理通信)。
- `9090` — (可选) 服务器健康度指标端点。

## 3. 本地启动隧道 (开发电脑使用)
暴露本地开发中的 React 服务器：
```bash
./qtun -server 您的VPS_IP:4443 -sub react -local localhost:3000
```
*(结果：通过 `https://react.tunnels.yourdomain.com` 开启访问)*

带基础密码保护的安全模式：
```bash
./qtun -server 您的VPS_IP:4443 -sub admin-panel -local localhost:8080 -auth "admin:12345"
```
