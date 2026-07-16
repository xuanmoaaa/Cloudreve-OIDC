# OIDC Proxy —— 为你的 Cloudreve 加上彩虹聚合登录

将[彩虹聚合登录](https://login.mapay.cn/)（QQ / 微信 / 支付宝）包装成标准 **OpenID Connect (OIDC)** 协议的中转服务，使支持 OIDC 的第三方应用（如 **Cloudreve**、**NextCloud**、**GitLab**、**Grafana** 等）能够接入国内社交平台登录。

```
第三方应用 <── OIDC ──> 本中转站 <── 彩虹 OAuth ──> QQ / 微信 / 支付宝
```

## ✨ 功能特性

- 完整实现 OIDC 核心端点：`authorize` / `token` / `userinfo` / `.well-known`
- 优雅的登录通道选择页（QQ / 微信 / 支付宝图标）
- **Web 管理后台**：客户端管理、系统设置、登录日志查看
- 配置持久化到 **SQLite**，无需外部数据库，开箱即用
- RSA 密钥对自动生成，id_token 使用 **RS256** 签名
- **零前端依赖**，纯 HTML + 内联 CSS，单二进制部署
- Docker 支持，多阶段构建，镜像极小

## 📦 前置准备

- 一个已搭建好的 **Cloudreve v4** 网盘（或其他支持 OIDC 的应用）
- [彩虹聚合登录](https://login.mapay.cn/) 账号，并创建应用获取 **APP ID** 和 **APP KEY**

## 🚀 快速开始

### 直接运行

```bash
# 编译
go build -o oidc-proxy .

# 启动（自动生成 RSA 密钥对，自动创建 SQLite 数据库）
./oidc-proxy -db data.db
```

首次启动后访问 `http://localhost:8443/admin/login`，使用默认账号登录：

- 用户名：`admin`
- 密码：`admin123`

在管理后台「系统设置」中填入彩虹聚合登录的 **APP ID** 和 **APP KEY**，重启服务即可。

### Docker 部署

```bash
docker build -t oidc-proxy .
docker run -d -p 8443:8443 -v ./data:/app/keys -v ./data:/app oidc-proxy
```

## 📋 OIDC 端点

| 端点 | 方法 | 说明 |
|------|------|------|
| `/.well-known/openid-configuration` | GET | OIDC 发现文档 |
| `/.well-known/jwks.json` | GET | JWKS 公钥 |
| `/oidc/authorize` | GET | 授权端点（展示通道选择页） |
| `/oidc/callback` | GET | 彩虹登录回调 |
| `/oidc/token` | POST | 用 code 换取 id_token |
| `/oidc/userinfo` | GET | 用户信息 |

### 授权流程

1. 第三方应用重定向到 `GET /oidc/authorize?client_id=xxx&redirect_uri=xxx&response_type=code&state=xxx&nonce=xxx`
2. 用户看到通道选择页，选择 QQ / 微信 / 支付宝
3. 跳转至彩虹聚合登录完成授权
4. 回调后生成一次性 `code`，302 跳回第三方应用的 `redirect_uri`
5. 第三方应用 `POST /oidc/token` 用 `code` 换取 `id_token`（JWT / RS256）

## 🖥️ 管理后台

访问 `/admin/login` 登录后，提供三个功能 Tab：

| Tab | 功能 |
|-----|------|
| 仪表盘 | 运行状态、彩虹 API 连通性、最近登录日志 |
| 客户端 | 创建 / 删除 OIDC 客户端，查看 Client ID & Secret |
| 系统设置 | 修改 Issuer、管理员账号、彩虹 API 参数 |

所有配置实时保存到 SQLite，修改后需重启服务生效。

## 📁 项目结构

```
oidc-proxy/
├── main.go                     # 入口：路由注册 + 启动
├── internal/
│   ├── model/types.go          # 数据结构定义
│   ├── crypto/jwt.go           # RSA 密钥管理 + JWT 签名
│   ├── store/sqlite.go         # SQLite 存储层
│   ├── rainbow/client.go       # 彩虹聚合登录 API 封装
│   ├── handler/
│   │   ├── authorize.go        # OIDC 授权 + 通道选择页
│   │   ├── callback.go         # 彩虹回调处理
│   │   ├── token.go            # Token 端点
│   │   ├── userinfo.go         # UserInfo 端点
│   │   ├── discovery.go        # OIDC 发现文档
│   │   ├── jwks.go             # JWKS 端点
│   │   ├── admin.go            # Web 管理后台
│   │   ├── assets.go           # 嵌入静态资源（登录图标）
│   │   └── handler.go          # Handler 结构体定义
│   └── middleware/auth.go      # 管理员认证中间件
├── picture/                    # 登录方式图标
│   ├── qq.png
│   ├── wx.png
│   └── alipay.png
├── Dockerfile
├── go.mod
└── README.md
```

## ⚙️ 配置项

所有配置通过管理后台在线修改，存储在 SQLite `settings` 表中：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `issuer` | OIDC 签发者 URL | `https://oidc.example.com` |
| `listen` | 监听地址 | `:8443` |
| `admin_username` | 管理员用户名 | `admin` |
| `admin_password` | 管理员密码 | `admin123` |
| `rainbow_app_id` | 彩虹聚合登录 APP ID | （需填写） |
| `rainbow_app_key` | 彩虹聚合登录 APP KEY | （需填写） |
| `rainbow_connect_url` | 彩虹 connect.php 地址 | `https://login.mapay.cn/connect.php` |

## 🔧 技术栈

- **Go 1.21+** · 标准库 `net/http`
- [go-jose/go-jose v3](https://github.com/go-jose/go-jose) — JWT 签名 / JWKS
- [modernc.org/sqlite](https://modernc.org/sqlite) — 纯 Go 实现的 SQLite 驱动
- 纯 HTML + 内联 CSS · 零前端依赖

## 📝 第三方应用接入示例（NextCloud）

在 NextCloud 配置 `config.php` 中添加：

```php
'oidc_login_provider' => 'https://your-domain.com',
'oidc_login_client_id' => 'client_xxxxxxxxxxxx',
'oidc_login_client_secret' => 'secret_xxxxxxxxxxxxxxxx',
```

## 📬 联系我

- Telegram: [@xuanmo1314](https://t.me/xuanmo1314)
- QQ: 1557275609

## 📄 License

MIT
