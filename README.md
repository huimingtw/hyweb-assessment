# hyweb-assessment

RESTful API built with Go + Gin + MySQL.

## Prerequisites

- Go 1.22+
- Docker + Docker Compose
- [swag](https://github.com/swaggo/swag) CLI (for Swagger regeneration): `go install github.com/swaggo/swag/cmd/swag@latest`

---

## Quick Start (Docker)

```bash
cp .env.example .env
# Edit .env with your values

make docker-up
```

The API will be available at `http://localhost:8080`.  
Swagger UI: `http://localhost:8080/swagger/index.html`

---

## Local Setup (without Docker)

1. Copy and configure env vars:
   ```bash
   cp .env.example .env
   ```

2. Export variables (or use a tool like `direnv`):
   ```bash
   export PORT=8080
   export DB_USER=appuser
   export DB_PASS=secret
   export DB_NAME=hyweb
   export DB_HOST=localhost:3306
   export JWT_SECRET=your-secret
   export WEATHER_API_KEY=CWA-your-key   # optional; weather fetch skipped if empty
   ```

3. Apply the schema to your local MySQL:
   ```bash
   mysql -u root -p hyweb < db/schema.sql
   ```

4. Run the server:
   ```bash
   make dev
   ```

---

## Make Commands

| Command | Description |
|---------|-------------|
| `make dev` | Run with `go run main.go` |
| `make build` | Build binary to `bin/server` |
| `make swag` | Regenerate Swagger docs |
| `make docker-up` | Build and start containers |
| `make docker-down` | Stop and remove containers |

---

## Regenerate Swagger Docs

```bash
make swag
```

---

## API Endpoints

### Auth

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/register` | No | Register a new user |
| POST | `/api/v1/auth/login` | No | Login, returns JWT |

### User

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| PUT | `/api/v1/user/password` | Bearer JWT | Change password |

### Weather (bonus)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/weather` | No | Today's New Taipei City weather (public) |
| GET | `/api/v1/weather/me` | Bearer JWT | Today's New Taipei City weather (protected) |

---

## Example Requests

### Register
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret123"}'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret123"}'
```

### Change Password
```bash
curl -X PUT http://localhost:8080/api/v1/user/password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"old_password":"secret123","new_password":"newpass456"}'
```

### Get Weather (public)
```bash
curl http://localhost:8080/api/v1/weather
```

### Get Weather (protected)
```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/weather/me
```
