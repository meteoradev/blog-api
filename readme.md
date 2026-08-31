# BlogApi – Blog REST API with JWT Auth & Redis Caching

BlogApi is a production-ready REST API for managing blog posts and users. It supports user registration, JWT authentication, CRUD operations for posts with author validation, and Redis-powered caching with cache-aside invalidation. Built with Clean Architecture – fully typed, tested, and containerized.

---

## Killer Features

- **User Management** – Register, login, and authenticate via JWT (HS256, 1h access tokens). Passwords are hashed with bcrypt.
- **Post CRUD** – Create, read, update, and delete posts. Update and delete are protected: only the author can modify or remove their own posts (verified at the SQL level).
- **Redis Caching** – Cache-Aside pattern with 10-minute TTL. Posts are cached by ID and title; cache is invalidated on update and delete.
- **Rate Limiting** – Redis-based sliding window limiter: 5 RPM for public endpoints, 10 RPM for protected endpoints.
- **Quality First** – Integration tests with testcontainers, unit tests for domain and services, structured zerolog logging with request IDs.

---

## Tech Stack

![Go](https://img.shields.io/badge/Go-1.26-blue?logo=go)
![Chi Router](https://img.shields.io/badge/Chi-v5-green?logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-blue?logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-8-alpine?logo=redis)
![Docker](https://img.shields.io/badge/Docker-✓-blue?logo=docker)
![JWT](https://img.shields.io/badge/JWT-✓-purple?logo=jsonwebtokens)
![Bcrypt](https://img.shields.io/badge/Bcrypt-✓-orange)
![Zerolog](https://img.shields.io/badge/Zerolog-✓-lightgrey)
![Swagger](https://img.shields.io/badge/Swagger-2.0-yellow?logo=swagger)

---

## Quick Start

### Prerequisites

- [Git](https://git-scm.com/)
- [Docker](https://docker.com/) & [Docker Compose](https://docs.docker.com/compose/)

### Steps

```bash
# 1. Clone this repository:
git clone https://github.com/meteoradev/BlogApi.git
cd BlogApi

# 2. Setup .env file:
cp .env.example .env
# edit .env file with your data

# 3. Run the project:
docker compose up -d

# 4. After all, delete created network and containers:
docker compose down
```

### After startup, the API will be available at:

- **Main API:** http://localhost:8080
- **Swagger UI:** http://localhost:8080/swagger/
- **Swagger JSON:** http://localhost:8080/swagger/doc.json

---

## API Endpoints

### Auth

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/register` | ❌ | Register a new user (username, email, password). Returns 409 if email already exists. |
| POST | `/login` | ❌ | Authenticate with email and password. Returns a JWT access token (1h). |

### Posts

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/posts/` | ✅ | Create a new post (title ≤ 100 chars, content ≤ 1000 chars). |
| GET | `/posts/id/{postID}` | ❌ | Get a post by ID. Public, cached (10 min TTL). |
| GET | `/posts/title/{title}` | ❌ | Get a post by title. Public, cached (10 min TTL). |
| PUT | `/posts/{postID}` | ✅ | Update a post. Only the author can update. |
| DELETE | `/posts/{postID}` | ✅ | Delete a post. Only the author can delete. |

### Users

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/users/id/{userID}` | ✅ | Get user by ID. |
| GET | `/users/email/{email}` | ✅ | Get user by email (URL-encoded). |
| PUT | `/users/{userID}` | ✅ | Update user. Users can only update their own profile. |
| DELETE | `/users/{userID}` | ✅ | Delete user. Users can only delete their own profile. |

**All endpoints that require authentication expect a valid JWT token in the `Authorization: Bearer <token>` header.**

---

## Database Migrations

Migrations are managed with **golang-migrate**. To create or apply migrations:

```bash
# Create a new migration
migrate create -ext sql -dir migrations -seq "describe_your_changes"

# Apply pending migrations (local)
migrate -path migrations -database "postgres://user:pass@localhost:5432/db?sslmode=disable" up

# Apply pending migrations (Docker)
docker compose exec api migrate -path migrations -database "$DATABASE_URL" up
```

---

## Security Considerations

- **Passwords** are hashed with `bcrypt` (default cost). Never stored in plaintext.
- **JWT tokens** use HMAC-SHA256. Access tokens expire after 1 hour. No refresh tokens are implemented.
- **Authorisation** is enforced at the SQL level: `UPDATE posts SET ... WHERE id = $1 AND user_id = $2` ensures users can only modify their own posts.
- **Rate limiting** is applied via Redis sliding window to prevent brute-force and abuse.
- **SQL injection** is prevented by using parameterized queries through `sqlx`.

---
