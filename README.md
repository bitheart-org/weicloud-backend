# weicloud-backend

KVM 虚拟化管理平台后端服务（Go + Gin + GORM + PostgreSQL + Incus SDK）。

## Quick Start

```bash
make tidy
make dev
```

默认启动后可访问健康检查：

```bash
curl http://localhost:8080/health
```

## 环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `SERVER_ADDR` | `:8080` | HTTP 监听地址 |
| `DB_HOST` | `127.0.0.1` | PostgreSQL 地址 |
| `DB_PORT` | `5432` | PostgreSQL 端口 |
| `DB_USER` | `weicloud` | 数据库用户名 |
| `DB_PASSWORD` | `weicloud` | 数据库密码 |
| `DB_NAME` | `weicloud` | 数据库名 |
| `DB_SSLMODE` | `disable` | PostgreSQL SSL 模式 |
| `JWT_SECRET` | `change-me` | JWT 签名密钥 |
| `JWT_EXPIRE_HOURS` | `24` | Token 过期小时数 |
| `ADMIN_USERNAME` | `admin` | 初始管理员用户名 |
| `ADMIN_PASSWORD` | `Admin@123456` | 初始管理员密码 |
| `ADMIN_DISPLAY_NAME` | `Administrator` | 初始管理员显示名 |

## 主要接口

1. 认证：`/api/auth/login`、`/api/auth/me`
2. 管理端：`/api/admin/users`、`/api/admin/hosts`、`/api/admin/vms`、`/api/admin/dashboard`、`/api/admin/logs`
3. 用户端：`/api/user/vms`、`/api/user/vms/:id/{start|stop|reboot|password|resource}`
