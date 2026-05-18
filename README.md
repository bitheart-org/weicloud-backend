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

