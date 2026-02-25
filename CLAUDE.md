# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Start the PostgreSQL database
docker-compose up -d

# Run the application (requires .env file and running DB)
go run ./cmd/web/

# Build the binary
go build ./cmd/web/

# Run all tests
go test ./...

# Run a single test
go test ./internal/... -run TestName

# Regenerate ent ORM code after schema changes
go generate ./ent/
```

## Architecture

This is a Go web application for church administration. The entry point is `cmd/web/main.go`.

### Dependency Injection

The `application` struct in `cmd/web/app.go` acts as the DI container, holding the ent client, session manager, template cache, form decoder, and model clients. It is initialized once in `main.go` and its methods are the HTTP handlers and middleware.

### Request Flow

```
HTTP Request
  → alice standard middleware (recoverPanic → logRequest → commonHeaders)
  → alice dynamic middleware (sessionManager.LoadAndSave → noSurf CSRF)
  → [protected routes: requireAuthentication]
  → handler (app.*) in handler.go
  → model method in internal/models/
  → ent client in ent/
```

Routes are defined in `cmd/web/routes.go` using `net/http.ServeMux` and `justinas/alice` for middleware chaining.

### Database Layer (EntGo ORM)

- Schema source of truth: `ent/schema/*.go` — **only these files are manually edited**
- All other files in `ent/` are auto-generated via `go generate ./ent/`
- On startup, `connectDB` in `helpers.go` auto-migrates the schema via `client.Schema.Create(..., migrate.WithGlobalUniqueID(true))`
- Two separate DB connections are opened: one `*ent.Client` for queries, one `*sql.DB` for the SCS session store (`scs/postgresstore`)

### Data Model Relationships

- **Church** — hierarchical (headquarters → branch → mission) via self-referential `parent_id`; owns users, departments, events, finances, and invitations
- **User** — belongs to one church; has one contact; has a role enum; tracks finance records and sent/accepted invitations
- **Invitation** — token-based (JWT signed with `JWT_SECRET`); links a church, an inviter user, and an accepted user

### Frontend

- Go HTML templates in `ui/html/` using a `base.gohtml` layout with pages and partials
- All UI assets (templates + static files) are embedded into the binary via `//go:embed` in `ui/efs.go`
- Templates are pre-compiled into a cache map (`map[string]*template.Template`) at startup in `templates.go`
- CSS: [Bulma](https://bulma.io/) framework

### Validation Pattern

DTOs in `internal/models/dtos.go` embed `validator.Validator` (from `internal/validator/validators.go`). Handlers call `dto.CheckField(...)` and then `dto.Valid()` before passing to a model method. Models return a `ModelResponse{Data any, Error error}` struct.

### Email

`cmd/web/mail.go` provides `SendEmail()` which renders an HTML template and sends via SMTP. Configuration is read from env vars at send time (`MAIL_HOST`, `MAIL_PORT`, `MAIL_FROM`, `MAIL_USERNAME`, `MAIL_PASSWORD`).

## Environment Variables

The app loads `.env` from the working directory at startup (via `godotenv`). Required variables:

| Variable | Purpose |
|---|---|
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_NAME`, `DB_PASS`, `DB_SSL_MODE` | PostgreSQL connection |
| `JWT_SECRET` | Signs/verifies invitation tokens |
| `MAIL_HOST`, `MAIL_PORT`, `MAIL_FROM`, `MAIL_USERNAME`, `MAIL_PASSWORD` | SMTP email |

The Docker Compose DB uses port `5435` externally (mapped to `5432` internally).
