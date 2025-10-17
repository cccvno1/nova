# Nova RBAC System

A Go-based Role-Based Access Control (RBAC) system with hierarchical permission levels, featuring comprehensive user management, permission management, audit logging, and task scheduling capabilities.

## 🌟 Features

### Core Features
- **Hierarchical Role System**: Role levels (1-100) for fine-grained permission control
- **Row-Level Security (RLS)**: Automatic filtering based on user permission levels
- **JWT Authentication**: Secure token-based authentication with refresh tokens
- **Permission Management**: Flexible permission assignment with diff preview API
- **Audit Logging**: Comprehensive audit trail for all operations
- **Multi-tenancy**: Domain-based isolation for multi-tenant scenarios

### Advanced Features
- **Casbin Integration**: Policy-based access control
- **Rate Limiting**: Token bucket and sliding window algorithms
- **Task Scheduling**: Cron-based task scheduler
- **File Storage**: Local file storage management
- **Redis Caching**: Performance optimization with Redis
- **Middleware Stack**: CORS, recovery, logging, and custom middleware support

## 🏗️ Architecture

```
nova/
├── cmd/server/          # Application entry point
├── internal/            # Private application code
│   ├── handler/        # HTTP handlers
│   ├── model/          # Domain models
│   ├── repository/     # Data access layer
│   ├── router/         # Route definitions
│   ├── server/         # Server setup
│   └── service/        # Business logic
├── pkg/                # Public libraries
│   ├── auth/           # JWT & token blacklist
│   ├── cache/          # Redis caching
│   ├── casbin/         # Casbin enforcer wrapper
│   ├── config/         # Configuration management
│   ├── database/       # Database connection & GORM
│   ├── errors/         # Error handling
│   ├── logger/         # Structured logging
│   ├── middleware/     # HTTP middleware
│   ├── queue/          # Task queue
│   ├── ratelimit/      # Rate limiting
│   ├── response/       # HTTP response helpers
│   ├── scheduler/      # Task scheduler
│   ├── storage/        # File storage
│   └── validator/      # Request validation
├── web/                # Vue 3 frontend
├── configs/            # Configuration files
├── scripts/            # Database migrations & scripts
└── docs/               # Documentation
```

## 🚀 Quick Start

### Prerequisites
- Go 1.24+
- PostgreSQL 16+
- Redis 7+
- Docker & Docker Compose (optional)

### Using Docker (Recommended)

```bash
# Start all services
docker-compose -f docker-compose.test.yml up -d

# View logs
docker logs -f nova-server-test

# Access the application
# Backend API: http://localhost:8080
# Frontend: http://localhost:5173 (if running web separately)
```

### Manual Setup

1. **Configure Database**
```bash
cp configs/config.yaml.example configs/config.yaml
# Edit configs/config.yaml with your database settings
```

2. **Run Migrations**
```bash
psql -U postgres -d nova -f scripts/migrations/001_init_rbac.sql
psql -U postgres -d nova -f scripts/migrations/002_add_role_level.sql
```

3. **Start Server**
```bash
go run cmd/server/main.go
```

4. **Initialize Default Data**
```bash
cd scripts/migrations
go run seed_rbac_data.go
```

## 📊 Database Schema

### Core Tables
- **users**: User accounts with password hashing
- **roles**: Role definitions with hierarchy levels
- **permissions**: Permission definitions
- **user_roles**: User-role assignments with audit info
- **role_permissions**: Role-permission mappings
- **casbin_rule**: Casbin policy storage
- **audit_logs**: Comprehensive operation audit trail

## 🔐 Security Features

### Role Level System
- Each role has a level (1-100)
- Users can only manage roles with lower levels
- Row-level security filters query results automatically
- Prevents privilege escalation attacks

### Authentication & Authorization
- JWT-based authentication with access & refresh tokens
- Token blacklist for logout
- Per-endpoint permission checks
- Domain-based multi-tenancy

### Rate Limiting
- Token bucket algorithm
- Sliding window counter
- Configurable limits per endpoint

## 🐛 Known Issues

### Critical Issue: Data Inconsistency Between user_roles and casbin_rule

**Problem**: The system has a design flaw where user-role assignments are stored in two places:
1. `user_roles` table (business layer with audit info)
2. `casbin_rule` table (permission engine)

**Current Behavior**:
- Write operations only update `user_roles` table
- Read operations query `casbin_rule` table
- Casbin stores role **names** but code expects role **IDs**
- This causes `GetUserRoles` to return empty lists
- Results in permission checks failing and empty role lists in UI

**Example**:
```sql
-- user_roles table
user_id=1, role_id=2, domain='default'  ✓

-- casbin_rule table  
v0='1', v1='super_admin', v2='default'  ✓
v0='1', v1='admin', v2='default'  ✗ (missing)

-- Additionally, code tries to parse "super_admin" as uint, which fails
```

**Workaround**: Until fixed, you can manually sync data or query `user_roles` directly.

**Fix Options**:
- **Option A**: Use Casbin as single source of truth (store role IDs, not names)
- **Option B**: Use RBAC tables as single source, make Casbin optional
- **Option C**: Implement proper two-way sync in `AssignRolesToUser`

See [docs/12-RBAC权限管理重构方案.md](docs/12-RBAC权限管理重构方案.md) for detailed analysis.

## 🔧 Configuration

### Environment Variables
```yaml
# configs/config.yaml
server:
  port: 8080
  mode: debug  # debug, release, test

db:
  driver: postgres
  host: localhost
  port: 5432
  database: nova
  username: postgres
  password: postgres

redis:
  host: localhost
  port: 6379
  db: 0

jwt:
  secret: your-secret-key
  access_token_duration: 2h
  refresh_token_duration: 168h
```

## 📝 API Documentation

### Authentication
```http
POST /api/v1/auth/login
POST /api/v1/auth/logout
POST /api/v1/auth/refresh
```

### User Management
```http
GET    /api/v1/users
POST   /api/v1/users
GET    /api/v1/users/:id
PUT    /api/v1/users/:id
DELETE /api/v1/users/:id
```

### Role Management
```http
GET    /api/v1/roles
POST   /api/v1/roles
GET    /api/v1/roles/:id
PUT    /api/v1/roles/:id
DELETE /api/v1/roles/:id

# Permission assignment with diff preview
POST   /api/v1/roles/:id/permissions/update
```

### Permission Management
```http
GET    /api/v1/permissions
POST   /api/v1/permissions
GET    /api/v1/permissions/tree
```

Default credentials:
- Username: `admin`
- Password: `admin123`

## 🧪 Testing

```bash
# Run tests
make test

# Run with coverage
make test-coverage

# Test specific package
go test ./internal/service/...
```

## 📚 Documentation

Detailed documentation is available in the `docs/` directory:

- [00-导读.md](docs/00-导读.md) - Project overview
- [01-启动流程.md](docs/01-启动流程.md) - Startup process
- [06-RBAC 权限模块.md](docs/06-RBAC%20权限模块.md) - RBAC module
- [11-安全加固实施总结.md](docs/11-安全加固实施总结.md) - Security hardening
- [12-RBAC权限管理重构方案.md](docs/12-RBAC权限管理重构方案.md) - Refactoring plan

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- [Echo](https://echo.labstack.com/) - High performance web framework
- [GORM](https://gorm.io/) - ORM library
- [Casbin](https://casbin.org/) - Authorization library
- [Vue 3](https://vuejs.org/) - Progressive JavaScript framework
- [Element Plus](https://element-plus.org/) - Vue 3 component library

## 📧 Contact

Chen Chi - [@cccvno1](https://github.com/cccvno1)

Project Link: [https://github.com/cccvno1/nova](https://github.com/cccvno1/nova)

---

**Note**: This project is currently in active development. The known data inconsistency issue between `user_roles` and `casbin_rule` tables needs to be resolved before production use. See the Known Issues section above for details.
