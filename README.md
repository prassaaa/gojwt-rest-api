<div align="center">

# Go JWT REST API

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![MySQL](https://img.shields.io/badge/MySQL-5.7+-4479A1?style=for-the-badge&logo=mysql&logoColor=white)
![JWT](https://img.shields.io/badge/JWT-Authentication-000000?style=for-the-badge&logo=jsonwebtokens&logoColor=white)

[![Gin Framework](https://img.shields.io/badge/Gin-Framework-00ADD8?style=flat-square&logo=go)](https://gin-gonic.com/)
[![GORM](https://img.shields.io/badge/GORM-ORM-00ADD8?style=flat-square)](https://gorm.io/)

RESTful API dengan JWT Authentication yang dibangun menggunakan Go, mengikuti Clean Architecture dan best practices.

</div>

---

## Fitur

- **Authentication & Authorization**
  - Register dan Login dengan JWT
  - **Refresh Token Mechanism** dengan automatic token rotation
  - **Token Revocation & Blacklisting** untuk logout
  - Token reuse detection untuk keamanan lebih baik
  - Password hashing menggunakan bcrypt
  - Protected routes dengan JWT middleware
  - Short-lived access tokens (15 menit) & long-lived refresh tokens (7 hari)

- **User Self-Service**
  - User dapat mengelola profil sendiri
  - Change password dengan verifikasi password lama
  - Update profile (name & email)
  - Get own profile

- **CRUD Operations**
  - User management (Create, Read, Update, Delete)
  - Pagination dan filtering
  - Search functionality

- **Security & Performance**
  - **Refresh token rotation** untuk mencegah token reuse
  - **Token family tracking** untuk deteksi suspicious activity
  - Rate limiting
  - CORS middleware
  - Input validation
  - Graceful shutdown

- **Clean Architecture**
  - Separation of concerns (Domain, Repository, Service, Handler)
  - Interface-based design untuk testability
  - Dependency injection

## Tech Stack

- **Framework**: Gin
- **Database**: MySQL + GORM
- **Authentication**: JWT (golang-jwt/jwt)
- **Validation**: go-playground/validator
- **Password Hashing**: bcrypt

## Struktur Project

```
gojwt-rest-api/
├── cmd/
│   ├── api/             # Application entry point
│   └── tools/           # Tools (JWT secret generator)
├── internal/
│   ├── config/          # Configuration & database
│   ├── domain/          # Domain models & DTOs
│   ├── repository/      # Data access layer
│   ├── service/         # Business logic
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # Middleware (auth, rate limit, cors)
│   └── utils/           # Utilities (JWT, password)
├── pkg/                 # Public packages
│   ├── logger/
│   └── validator/
├── migrations/          # Database migrations
├── docs/                # Documentation
│   ├── HOT_RELOAD_GUIDE.md
│   └── JWT_SECRET_GUIDE.md          # Air hot reload config
├── .air.toml.example    # Air config template
├── .env.example         # Environment variables template
├── .gitignore
├── Makefile
└── README.md
```

## Setup & Installation

### Prerequisites

- Go 1.21+
- MySQL 5.7+

### Installation

1. Clone repository:
```bash
git clone https://github.com/prassaaa/gojwt-rest-api
cd gojwt-rest-api
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Generate JWT Secret:
```bash
# Menggunakan Go tool
go run cmd/tools/generate_secret.go

# Atau dengan make
make generate-secret

# Atau dengan OpenSSL
openssl rand -base64 32
```

4. Edit `.env` dan sesuaikan konfigurasi:
```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=gojwt_db
JWT_SECRET=your-super-secret-key
```
**⚠️ Ganti JWT_SECRET dengan hasil generate di step 3!**

5. Buat database MySQL:
```bash
mysql -u root -p
CREATE DATABASE gojwt_db;
```

6. Install dependencies:
```bash
go mod download
```

7. Build aplikasi:
```bash
go build -o bin/api cmd/api/main.go
```

8. Run aplikasi:
```bash
./bin/api
```

Atau langsung dengan:
```bash
go run cmd/api/main.go
```

Server akan berjalan di `http://localhost:8080`

## API Endpoints

### Health Check
```
GET /health
```

### Authentication (Public)

**Register**
```
POST /api/v1/auth/register
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123"
}
```

**Login**
```
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "password123"
}

Response:
{
  "success": true,
  "message": "login successful",
  "data": {
    "user": {...},
    "access_token": "eyJhbGc...",
    "refresh_token": "random_secure_token",
    "expires_in": 900,
    "token_type": "Bearer"
  }
}
```

**Refresh Token** (New!)
```
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "your_refresh_token_here"
}

Response:
{
  "success": true,
  "message": "token refreshed successfully",
  "data": {
    "access_token": "new_access_token",
    "refresh_token": "new_refresh_token",
    "expires_in": 900,
    "token_type": "Bearer"
  }
}
```

**Logout** (New!)
```
POST /api/v1/auth/logout
Authorization: Bearer <your-access-token>
Content-Type: application/json

{
  "refresh_token": "your_refresh_token_here"
}
```

### Profile (Protected - User Self-Service)

Semua endpoints di bawah memerlukan header:
```
Authorization: Bearer <your-jwt-token>
```

**Get Own Profile**
```
GET /api/v1/profile
```

**Update Own Profile**
```
PUT /api/v1/profile
Content-Type: application/json

{
  "name": "John Updated",
  "email": "johnupdated@example.com"
}
```

**Change Password**
```
PUT /api/v1/profile/password
Content-Type: application/json

{
  "old_password": "password123",
  "new_password": "newpassword123"
}
```

### Users (Protected - Admin Only)

Semua endpoints di bawah memerlukan header:
```
Authorization: Bearer <your-jwt-token>
```

**Get Profile** (deprecated - gunakan GET /api/v1/profile)
```
GET /api/v1/users/profile
```

**Get All Users (with pagination)**
```
GET /api/v1/users?page=1&page_size=10&search=john
```

**Get User by ID**
```
GET /api/v1/users/:id
```

**Update User**
```
PUT /api/v1/users/:id
Content-Type: application/json

{
  "name": "John Updated",
  "email": "johnupdated@example.com"
}
```

**Delete User**
```
DELETE /api/v1/users/:id
```

## Testing dengan cURL

### Register
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

### Logout
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

### Get Own Profile (dengan token)
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Test Refresh Token Feature
Jalankan automated test script:
```bash
./test_refresh_token.sh
```

### Update Own Profile
```bash
curl -X PUT http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Updated",
    "email": "johnupdated@example.com"
  }'
```

### Change Password
```bash
curl -X PUT http://localhost:8080/api/v1/profile/password \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "password123",
    "new_password": "newpassword123"
  }'
```

## Best Practices yang Diimplementasikan

1. **Clean Architecture**
   - Domain layer: Business entities
   - Repository layer: Data access abstraction
   - Service layer: Business logic
   - Handler layer: HTTP handling

2. **Interface-based Design**
   - Repository dan Service menggunakan interface
   - Memudahkan testing dan dependency injection

3. **Error Handling**
   - Consistent error responses
   - Custom error messages
   - Proper HTTP status codes

4. **Security**
   - Password hashing dengan bcrypt
   - JWT token authentication
   - Rate limiting untuk mencegah abuse
   - Input validation

5. **Configuration Management**
   - Environment variables
   - Default values
   - Configuration validation

6. **Database**
   - Connection pooling
   - Auto migrations
   - Proper indexing (email unique index)

7. **Middleware**
   - Authentication middleware
   - Rate limiting middleware
   - CORS middleware

8. **Graceful Shutdown**
   - Signal handling (SIGINT, SIGTERM)
   - Connection cleanup
   - Timeout context

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| SERVER_PORT | Server port | 8080 |
| SERVER_HOST | Server host | localhost |
| DB_HOST | Database host | localhost |
| DB_PORT | Database port | 3306 |
| DB_USER | Database user | root |
| DB_PASSWORD | Database password | - |
| DB_NAME | Database name | gojwt_db |
| JWT_SECRET | JWT secret key | - (required) |
| JWT_EXPIRATION_HOURS | Token expiration | 24 |
| RATE_LIMIT_REQUESTS | Rate limit requests | 100 |
| RATE_LIMIT_DURATION | Rate limit duration | 1m |
| APP_ENV | Environment | development |

## Development

### Run dengan hot reload:

Air sudah dikonfigurasi untuk hot reload. Setiap kali save file `.go`, aplikasi akan otomatis rebuild dan restart.

```bash
# Install Air (jika belum)
go install github.com/air-verse/air@latest

# Tambahkan ~/go/bin ke PATH (jika belum)
export PATH=$PATH:~/go/bin

# Run dengan Air
air

# Atau dengan Makefile
make dev
```

**Cara kerja**: Edit file `.go` → Save → Air auto reload → Test langsung!

Lihat dokumentasi lengkap: [docs/HOT_RELOAD_GUIDE.md](docs/HOT_RELOAD_GUIDE.md)

### Format code:
```bash
go fmt ./...
```

### Lint:
```bash
golangci-lint run
```

## Production Deployment

1. Set `APP_ENV=production` di environment
2. Use strong `JWT_SECRET`
3. Setup proper database credentials
4. Use reverse proxy (Nginx)
5. Enable HTTPS
6. Setup monitoring dan logging

## Documentation

- **[Refresh Token Guide](./docs/REFRESH_TOKEN.md)** - Comprehensive guide untuk refresh token mechanism, token rotation, dan security features
- **[JWT Secret Guide](./docs/JWT_SECRET_GUIDE.md)** - Panduan generate dan manage JWT secrets
- **[Hot Reload Guide](./docs/HOT_RELOAD_GUIDE.md)** - Setup Air untuk development dengan hot reload
