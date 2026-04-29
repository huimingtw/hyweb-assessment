# CLAUDE.md

Engineering spec and conventions for this project. Follow these exactly.

---

## Project Overview

A RESTful API backend built with Go + Gin + MySQL.

**Deliverables required:**
- All API responses in a unified envelope format (see below)
- User registration, login, change password
- JWT authentication on protected routes
- Dockerfile + docker-compose (app + MySQL)
- Swagger UI or Postman collection
- README with local setup instructions
- Bonus: Central Weather Bureau API integration (New Taipei City weather, scheduled every 24h)

---

## Tech Stack

- **Language**: Go
- **Framework**: Gin
- **Database**: MySQL
- **Auth**: JWT (`golang-jwt/jwt`)
- **Password hashing**: `bcrypt`
- **Migration**: `golang-migrate` or raw SQL `CREATE TABLE` statements provided in `docs/`
- **API Docs**: Swagger UI via `swaggo/swag`
- **Container**: Docker + docker-compose

---

## Project Structure

```
.
├── main.go
├── config/
│   └── config.go          # env-based config, no file persistence needed
├── handler/
│   ├── response.go        # Response struct, OK, Fail helpers
│   ├── handler.go         # Handler struct + constructor
│   ├── auth.go            # Register, Login
│   ├── user.go            # ChangePassword
│   └── weather.go         # GetWeather (public + protected)
├── service/
│   ├── errors.go          # sentinel errors
│   ├── auth_service.go
│   ├── user_service.go
│   └── weather_service.go
├── middleware/
│   ├── logger.go          # Request ID + structured logging
│   └── jwt.go             # JWT validation middleware
├── model/
│   ├── user.go
│   ├── weather.go
│   └── claims.go          # JWTClaims struct, shared by handler and middleware
├── router/
│   └── router.go
├── db/
│   ├── db.go
│   └── schema.sql         # CREATE TABLE statements for MySQL init
├── docs/                  # Swagger generated files
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

---

## Unified API Response Format

**Every** API endpoint must return this envelope. No exceptions.

```go
type Response struct {
    Success   bool        `json:"success"`
    Message   string      `json:"message"`
    Data      interface{} `json:"data"`
    Error     interface{} `json:"error"`
    Code      int         `json:"code"`
    Timestamp string      `json:"timestamp"` // RFC3339 UTC, e.g. "2025-09-12T07:58:43Z"
}
```

Helper functions to use everywhere:

```go
func OK(c *gin.Context, message string, data interface{})
func Fail(c *gin.Context, httpStatus int, message string, detail interface{})
```

`OK` sets `success: true`, `error: null`, `code: 200`.
`Fail` sets `success: false`, `data: null`, `code` = HTTP status.
`Timestamp` is always `time.Now().UTC().Format(time.RFC3339)`.

Never call `c.JSON` directly in handlers — always go through `OK` or `Fail`.

---

## Database Schema

### Users table

```sql
-- email is always stored lowercase (enforced by the service layer)
CREATE TABLE users (
    email    VARCHAR(50)  NOT NULL COLLATE utf8mb4_bin,
    password VARCHAR(255) NOT NULL,
    created  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### Weather table (bonus)

```sql
CREATE TABLE weather (
    id      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    city    VARCHAR(50)     NOT NULL,
    min_t   DECIMAL(4,1)    NOT NULL,
    max_t   DECIMAL(4,1)    NOT NULL,
    period  ENUM('AM','PM') NOT NULL,
    date    DATE            NOT NULL,
    created DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_city_period_date (city, period, date)
);
```

---

## API Endpoints

### Auth (no JWT required)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login, returns JWT |

### User (JWT required)

| Method | Path | Description |
|--------|------|-------------|
| PUT | `/api/v1/user/password` | Change password |

### Weather (bonus)

| Method | Path | JWT | Description |
|--------|------|-----|-------------|
| GET | `/api/v1/weather` | No | Public weather query |
| GET | `/api/v1/weather/me` | Yes | Protected weather query |

---

## Feature Specs

### POST /api/v1/auth/register
- Body: `{ "email": string, "password": string }`
- Validate: email format, password non-empty
- **Normalise email to lowercase** before any DB operation (service layer, not handler)
- Hash password with `bcrypt` (cost 12)
- If email already exists → `Fail` 409 `"email already exists"`
- On success → `OK` 200, `data: { "email": string, "created": string }`

### POST /api/v1/auth/login
- Body: `{ "email": string, "password": string }`
- **Normalise email to lowercase** before DB lookup
- Compare bcrypt hash; if either email or password is wrong → `Fail` 401 `"invalid credentials"` — do NOT reveal which field is wrong
- On success → `OK` 200, `data: { "token": string }`
- JWT payload must include: `email` (string, lowercase), `updated` (RFC3339 UTC string of the user's `updated` column)
- JWT expiry: 24 hours

### PUT /api/v1/user/password
- Header: `Authorization: Bearer <token>`
- Body: `{ "old_password": string, "new_password": string }`
- Verify old password against bcrypt hash; if wrong → `Fail` 400 `"old password is incorrect"`
- Hash new password with bcrypt (cost 12), update DB
- Existing tokens remain valid after password change (no blacklist required)
- On success → `OK` 200, `data: null`

### GET /api/v1/weather and GET /api/v1/weather/me
- Both return the same data; `/me` requires JWT, `/weather` is public
- Source: DB rows where `city = '新北市'` and `date = CURDATE()`
- AM = period stored from CWA `time[0]` (06:00–18:00), PM = `time[1]` (18:00–06:00)
- Response: `data: { "AM": { "min_t": float, "max_t": float }, "PM": { "min_t": float, "max_t": float } }`
- If no data yet: `"AM": null, "PM": null`

---

## JWT Middleware

```go
func JWTAuth(secret string) gin.HandlerFunc
```

- Reads `Authorization: Bearer <token>` header
- Parses and validates JWT signature + expiry
- On failure → `Fail` 401 `"unauthorized"` and `ctx.Abort()`
- On success → stores `email` claim in Gin context: `c.Set("email", claims.Email)`
- Handlers retrieve it with: `email := c.GetString("email")`
- `JWTClaims` struct is defined in `model/claims.go` and imported by both `handler` and `middleware` to avoid import cycles

---

## Observability

### Request ID
- Logger middleware reads `X-Request-Id` from incoming header; if absent, generates one with `crypto/rand` (8 bytes → 16-char hex)
- Writes the ID back to the response header
- Background goroutines must use `context.Background()`, never a request-scoped context

### Structured Logging
- Use `log/slog` exclusively. No `fmt.Println`, no `log.Print`
- Log output goes to `os.Stdout` only — no log files, no log rotation
- `slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})`
- Log levels: 5xx → `Error`, 4xx → `Warn`, 2xx/3xx → `Info`, internal state transitions → `Debug`
- Always include business keys as structured fields: `email`, `status`, etc.
- Logger middleware logs: `request_id`, `method`, `path`, `status`, `ip`, `latency`, `size`

---

## Dependency Injection

- All dependencies injected via constructors. No package-level globals.
- `Handler` struct receives db, logger, config, and services via `NewHandler()`
- Services receive their dependencies the same way
- `router.New(h *Handler, cfg *Config, logger *slog.Logger) *gin.Engine` — router receives handler, not the other way around
- Router must use `gin.New()` (not `gin.Default()`) and attach middleware explicitly: `gin.Recovery()`, then `RequestLogger`

---

## Graceful Shutdown

`main.go` must start the server in a goroutine and block on OS signal:

```go
srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}

go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        logger.Error("server error", "err", err)
        os.Exit(1)
    }
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := srv.Shutdown(ctx); err != nil {
    logger.Error("shutdown error", "err", err)
}
```

---

## Configuration

Load exclusively from environment variables (no config file needed for this project):

```go
type Config struct {
    Port          string
    DSN           string // built from DB_USER/DB_PASS/DB_NAME/DB_HOST
    JWTSecret     string
    WeatherAPIKey string // bonus, optional
}
```

DSN is constructed in `config.Load()` from individual vars:
- `DB_USER` (required)
- `DB_PASS` (required)
- `DB_NAME` (required)
- `DB_HOST` (optional, defaults to `localhost:3306`)

Fail fast on startup if `DB_USER`, `DB_PASS`, `DB_NAME`, or `JWT_SECRET` are missing.

`.env.example` must document every variable.

---

## Database

### MySQL conventions
- Use `database/sql` with `go-sql-driver/mysql`
- Blank import `_ "github.com/go-sql-driver/mysql"` in `db/db.go` only — importing it in multiple files causes a panic (double registration)
- After `sql.Open`, always call `db.Ping()` to verify connectivity — fail fast on startup if it fails
- Always propagate context: use `QueryRowContext`, `ExecContext`, `QueryContext`
- Use transactions (`BeginTx`) for multi-step writes; always `defer tx.Rollback()`
- Return `sql.ErrNoRows` transparently from the DB layer; handle it in the service layer

### Email casing
- **Email is always stored and queried in lowercase** — normalise with `strings.ToLower(email)` at the top of every service method that accepts an email parameter (Register, Login, ChangePassword)
- The `email` column uses `COLLATE utf8mb4_bin` (case-sensitive, binary) so the DB enforces exact-match on the already-lowercased value and prevents any mixed-case duplicates from slipping through

### Error mapping (service → handler)
- `sql.ErrNoRows` → 404
- Duplicate key (MySQL error 1062) → 409
- Other DB errors → 500, log actual error, return `"internal server error"`

---

## Error Handling

- Bcrypt cost: define `const BcryptCost = 12` in `service/auth_service.go` and reuse in `ChangePassword`

- Use sentinel errors for business rule violations:
  ```go
  var ErrInvalidCredentials = errors.New("invalid credentials")
  var ErrEmailExists        = errors.New("email already exists")
  var ErrWrongPassword      = errors.New("old password is incorrect")
  ```
- Handlers use `errors.Is()` to map sentinel errors to HTTP status codes
- Never expose raw DB error strings in responses

---

## Weather Scheduler (Bonus)

On startup: fetch immediately, then every 24 hours.

```go
go func() {
    weatherSvc.FetchAndStore(context.Background())
    for {
        time.Sleep(24 * time.Hour)
        weatherSvc.FetchAndStore(context.Background())
    }
}()
```

### CWA API

- **Endpoint**: `https://opendata.cwa.gov.tw/api/v1/rest/datastore/F-C0032-001`
- **Auth**: query param `Authorization=<key>` (NOT a header)
- **Key filter params**: `locationName=新北市`, `elementName=MinT`, `elementName=MaxT`
- **API key** from env var `WEATHER_API_KEY` — never hardcode
- If `WEATHER_API_KEY` is empty → log warning and return early (no crash)
- Use injected `*http.Client{Timeout: 10s}` — never `http.DefaultClient`

### Response parsing

F-C0032-001 returns a 36-hour forecast in 12-hour periods:

```json
{
  "records": {
    "location": [{
      "locationName": "新北市",
      "weatherElement": [
        { "elementName": "MinT", "time": [
            { "startTime": "2026-04-28 06:00:00", "parameter": { "parameterName": "22" } },
            { "startTime": "2026-04-28 18:00:00", "parameter": { "parameterName": "20" } }
          ]
        },
        { "elementName": "MaxT", "time": [ ... ] }
      ]
    }]
  }
}
```

- Filter `locationName == "新北市"`
- Build map: `elementName → []time`
- `time[0]` (06:00–18:00) → `period = 'AM'`
- `time[1]` (18:00–06:00) → `period = 'PM'`
- Temperature is in `parameter.parameterName` (string → parse to float64)

### Upsert

```sql
INSERT INTO weather (city, min_t, max_t, period, date)
VALUES ('新北市', ?, ?, ?, CURDATE())
ON DUPLICATE KEY UPDATE min_t = VALUES(min_t), max_t = VALUES(max_t), updated = NOW()
```

UNIQUE KEY is `(city, period, date)` — `date` column holds `CURDATE()`, not `DATE(created)`.

- Log success with `InfoContext`, failure with `ErrorContext`

### WeatherService constructor

```go
func NewWeatherService(db *sql.DB, logger *slog.Logger, apiKey string, client *http.Client) *WeatherService
```

### GetTodayWeather

```sql
SELECT period, min_t, max_t FROM weather
WHERE city = '新北市' AND date = CURDATE()
```

Returns `*model.WeatherResponse{AM: ..., PM: ...}` — pointer fields are `nil` if no row for that period.

---

## Docker

### Dockerfile
- Multi-stage build: `golang:1.25-alpine` builder → `alpine:latest` runner
- Runner image must include `ca-certificates` and `tzdata` (required for CWA HTTPS calls and time parsing)
- Copy only the compiled binary to the final image
- Expose the configured port

### docker-compose.yml
- Two services: `app` and `mysql`
- `app` depends on `mysql` with a health check
- All secrets via environment variables, sourced from `.env`
- MySQL data persisted via named volume

---

## API Documentation

Use `swaggo/swag`. Annotate all handlers with Swagger comments. Run `make swag` to generate `docs/`. Serve Swagger UI at `/swagger/index.html`.

---

## Make Commands

| Command | Description |
|---------|-------------|
| `make dev` | Start the server with `go run .` |
| `make build` | Build binary to `bin/server` |
| `make swag` | Regenerate Swagger docs (`swag init --parseDependency --parseInternal -g main.go --output docs`) |
| `make docker-up` | Build and start all containers |
| `make docker-down` | Stop and remove containers |



---

## README

Must include:
1. Prerequisites (Go version, Docker)
2. Local setup with docker-compose (single command)
3. Local setup without Docker (env vars, `go run`)
4. How to run `swag init` to regenerate docs
5. Example requests for every endpoint
