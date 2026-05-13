# E-Commerce API

A complete, production-patterned REST API written in Go. Built as a practice project covering JWT authentication, PostgreSQL, transactional order management, role-based access control, and clean layered architecture.

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Architecture](#architecture)
3. [Project Structure](#project-structure)
4. [Tech Stack](#tech-stack)
5. [Database Schema](#database-schema)
6. [Environment Variables](#environment-variables)
7. [API Reference](#api-reference)
   - [Auth](#auth)
   - [Users](#users)
   - [Categories](#categories)
   - [Products](#products)
   - [Cart](#cart)
   - [Orders](#orders)
   - [Reviews](#reviews)
8. [Authentication Flow](#authentication-flow)
9. [Role System](#role-system)
10. [Key Business Rules](#key-business-rules)
11. [Response Format](#response-format)
12. [Design Decisions](#design-decisions)

---

## Quick Start

```bash
# 1. Clone and enter the directory
cd ecommerce_application_api

# 2. Copy and configure environment
cp .env.example .env
# Edit .env — at minimum set DATABASE_URL

# 3. Create the database
make createdb   # runs: createdb ecommerce

# 4. Run migrations (creates all tables)
make migrate    # runs: psql $DATABASE_URL -f migrations/001_init.sql

# 5. (Optional) Load sample data
make seed       # runs: psql $DATABASE_URL -f migrations/002_seed.sql
                # Creates admin@example.com / Admin@123 + 5 categories + 5 products

# 6. Download Go dependencies (first time only)
make tidy

# 7. Start the server
make run        # listens on http://localhost:8080

# Health check
curl http://localhost:8080/api/v1/health
# → {"status":"ok"}
```

---

## Architecture

The app follows a classic **layered architecture** with unidirectional dependency flow:

```
HTTP Request
     │
     ▼
┌─────────────┐
│   Router    │  chi — defines all routes, applies middleware chains
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Middleware │  JWT auth, admin guard, CORS, logging, recovery
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Handlers   │  Parse request → validate → call repository → write response
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Repositories│  Interfaces + SQL implementations (pgx v5)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ PostgreSQL  │  Single pgxpool.Pool shared across all repos
└─────────────┘

Supporting packages (no DB knowledge):
  models  — all domain structs and request/response types
  utils   — JWT generation/validation, bcrypt wrappers, JSON response helpers
  config  — reads env vars, provides typed Config struct
```

**Key principle:** each layer only imports the layer directly below it. Handlers never touch `pgxpool`. Repositories never import `net/http`.

---

## Project Structure

```
ecommerce_application_api/
│
├── cmd/
│   └── api/
│       └── main.go              # Entry point — wires everything, starts HTTP server with graceful shutdown
│
├── internal/
│   ├── config/
│   │   └── config.go            # Reads env vars into a typed Config struct
│   │
│   ├── database/
│   │   └── db.go                # Creates and pings the pgxpool connection pool
│   │
│   ├── models/
│   │   └── models.go            # All domain types (User, Product, Order…) + all request/response types
│   │
│   ├── utils/
│   │   ├── jwt.go               # GenerateAccessToken, GenerateRefreshToken, ValidateToken
│   │   ├── password.go          # HashPassword (bcrypt), CheckPassword
│   │   └── response.go          # WriteSuccess, WriteError, WriteMessage helpers
│   │
│   ├── middleware/
│   │   ├── auth.go              # Authenticate (reads Bearer token), RequireAdmin, GetClaims helper
│   │   └── cors.go              # Permissive CORS for local development
│   │
│   ├── repository/
│   │   ├── user.go              # CRUD for users table
│   │   ├── category.go          # CRUD for categories table
│   │   ├── product.go           # CRUD + paginated list with filters for products
│   │   ├── cart.go              # GetOrCreate cart, add/update/remove/clear items
│   │   ├── order.go             # Create order (transactional), list, get, update status, cancel (transactional)
│   │   └── review.go            # CRUD for reviews, list by product with JOIN to users
│   │
│   ├── handlers/
│   │   ├── auth.go              # Register, Login, Refresh
│   │   ├── user.go              # GetMe, UpdateMe, ChangePassword
│   │   ├── category.go          # List, GetByID, Create, Update, Delete
│   │   ├── product.go           # List (paginated+filtered), GetByID, Create, Update, Delete
│   │   ├── cart.go              # GetCart, AddItem, UpdateItem, RemoveItem, ClearCart
│   │   ├── order.go             # Create, List, GetByID, UpdateStatus (admin), Cancel
│   │   └── review.go            # ListByProduct, Create, Update, Delete
│   │
│   └── router/
│       └── router.go            # Mounts all routes, applies middleware groups, wires handlers
│
├── migrations/
│   ├── 001_init.sql             # Creates all tables with constraints and indexes
│   └── 002_seed.sql             # Sample admin user, 5 categories, 5 products
│
├── cmd/api/main.go              # (see above)
├── go.mod                       # Module: ecommerce-api, Go 1.22
├── go.sum                       # Dependency checksums (auto-generated)
├── .env.example                 # Template for environment variables
├── Makefile                     # run, build, migrate, seed, tidy, createdb, dropdb
└── README.md                    # This file
```

---

## Tech Stack

| Concern | Library | Why |
|---------|---------|-----|
| HTTP router | `github.com/go-chi/chi/v5` | Idiomatic, lightweight, excellent middleware support |
| Database driver | `github.com/jackc/pgx/v5` | Native PostgreSQL driver, fast, no ORM overhead |
| JWT | `github.com/golang-jwt/jwt/v5` | Standard JWT implementation for Go |
| UUID | `github.com/google/uuid` | Generates and parses RFC 4122 UUIDs |
| Password hashing | `golang.org/x/crypto/bcrypt` | Industry-standard adaptive hashing |
| Env loading | `github.com/joho/godotenv` | Loads `.env` file into `os.Getenv` |

No ORM is used. All SQL is written by hand in the repository layer. This is intentional — it keeps queries explicit and easy to debug.

---

## Database Schema

```
users
  id            UUID PK
  email         VARCHAR(255) UNIQUE NOT NULL
  password_hash VARCHAR(255) NOT NULL
  name          VARCHAR(255) NOT NULL
  role          VARCHAR(50)  NOT NULL  DEFAULT 'customer'  -- 'customer' | 'admin'
  created_at    TIMESTAMPTZ
  updated_at    TIMESTAMPTZ

categories
  id            UUID PK
  name          VARCHAR(255) UNIQUE NOT NULL
  description   TEXT
  created_at    TIMESTAMPTZ
  updated_at    TIMESTAMPTZ

products
  id            UUID PK
  name          VARCHAR(255) NOT NULL
  description   TEXT
  price         NUMERIC(10,2)  CHECK (price >= 0)
  stock         INTEGER        CHECK (stock >= 0)
  category_id   UUID FK → categories(id)  ON DELETE SET NULL  (nullable)
  image_url     VARCHAR(500)
  created_at    TIMESTAMPTZ
  updated_at    TIMESTAMPTZ

carts
  id            UUID PK
  user_id       UUID UNIQUE FK → users(id)  ON DELETE CASCADE
  created_at    TIMESTAMPTZ
  updated_at    TIMESTAMPTZ

cart_items
  id            UUID PK
  cart_id       UUID FK → carts(id)     ON DELETE CASCADE
  product_id    UUID FK → products(id)  ON DELETE CASCADE
  quantity      INTEGER  CHECK (quantity > 0)
  created_at    TIMESTAMPTZ
  updated_at    TIMESTAMPTZ
  UNIQUE(cart_id, product_id)   -- one row per product per cart

orders
  id               UUID PK
  user_id          UUID FK → users(id)  ON DELETE RESTRICT
  status           VARCHAR(50)   -- pending | confirmed | shipped | delivered | cancelled
  total_amount     NUMERIC(10,2)
  shipping_address TEXT NOT NULL
  notes            TEXT
  created_at       TIMESTAMPTZ
  updated_at       TIMESTAMPTZ

order_items
  id          UUID PK
  order_id    UUID FK → orders(id)    ON DELETE CASCADE
  product_id  UUID FK → products(id)  ON DELETE RESTRICT
  quantity    INTEGER  CHECK (quantity > 0)
  price       NUMERIC(10,2)   -- price at time of purchase (snapshot)
  created_at  TIMESTAMPTZ

reviews
  id          UUID PK
  user_id     UUID FK → users(id)    ON DELETE CASCADE
  product_id  UUID FK → products(id) ON DELETE CASCADE
  rating      INTEGER  CHECK (rating >= 1 AND rating <= 5)
  comment     TEXT
  created_at  TIMESTAMPTZ
  updated_at  TIMESTAMPTZ
  UNIQUE(user_id, product_id)   -- one review per user per product
```

**Indexes:** `products.category_id`, `orders.user_id`, `orders.status`, `reviews.product_id`, full-text GIN index on `products.name`.

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://postgres:password@localhost:5432/ecommerce?sslmode=disable` | PostgreSQL connection string |
| `PORT` | `8080` | HTTP listen port |
| `JWT_SECRET` | *(weak default)* | Secret for signing access tokens — **change in production** |
| `JWT_EXPIRY_HOURS` | `24` | Access token lifetime in hours |
| `REFRESH_SECRET` | *(weak default)* | Secret for signing refresh tokens — **change in production** |
| `REFRESH_EXPIRY_DAYS` | `7` | Refresh token lifetime in days |

---

## API Reference

All endpoints are prefixed with `/api/v1`.

All responses wrap data in a standard envelope (see [Response Format](#response-format)).

**Auth legend:** `—` = public, `user` = any authenticated user, `admin` = admin role only.

---

### Auth

#### `POST /auth/register`

Create a new customer account. Returns tokens on success.

**Request body:**
```json
{
  "email": "alice@example.com",
  "password": "secret123",
  "name": "Alice"
}
```

**Success `201`:**
```json
{
  "success": true,
  "data": {
    "access_token": "<jwt>",
    "refresh_token": "<jwt>",
    "user": { "id": "...", "email": "alice@example.com", "name": "Alice", "role": "customer", ... }
  }
}
```

**Errors:** `400` missing fields / password < 6 chars, `409` email already registered.

---

#### `POST /auth/login`

**Request body:**
```json
{ "email": "alice@example.com", "password": "secret123" }
```

**Success `200`:** same shape as register.

**Errors:** `401` wrong email or password.

---

#### `POST /auth/refresh`

Exchange a valid refresh token for a new access token.

**Request body:**
```json
{ "refresh_token": "<refresh_jwt>" }
```

**Success `200`:**
```json
{ "success": true, "data": { "access_token": "<new_jwt>" } }
```

**Errors:** `401` invalid or expired refresh token.

---

### Users

All user endpoints require `Authorization: Bearer <access_token>`.

#### `GET /users/me`

Returns the authenticated user's profile.

**Success `200`:**
```json
{
  "success": true,
  "data": { "id": "...", "email": "...", "name": "...", "role": "customer", "created_at": "..." }
}
```

---

#### `PUT /users/me`

Update name and/or email. Provide only the fields you want to change.

**Request body:**
```json
{ "name": "Alice Smith", "email": "alice.smith@example.com" }
```

**Errors:** `400` both fields empty, `409` new email already in use.

---

#### `PUT /users/me/password`

**Request body:**
```json
{ "current_password": "old", "new_password": "new123" }
```

**Success `200`:** `{ "success": true, "message": "password updated successfully" }`

**Errors:** `401` current password wrong, `400` new password < 6 chars.

---

### Categories

#### `GET /categories` — public

Returns all categories sorted by name.

```json
{ "success": true, "data": [ { "id": "...", "name": "Electronics", "description": "..." } ] }
```

---

#### `GET /categories/{id}` — public

Single category by UUID. `404` if not found.

---

#### `POST /categories` — admin

```json
{ "name": "Toys", "description": "Optional description" }
```

`201` on success. `409` if name already exists.

---

#### `PUT /categories/{id}` — admin

Partial update — only provided fields are changed.

```json
{ "name": "Toys & Games", "description": "Updated description" }
```

---

#### `DELETE /categories/{id}` — admin

Deletes the category. Products in this category have their `category_id` set to `NULL` (via `ON DELETE SET NULL`).

---

### Products

#### `GET /products` — public

Paginated list with optional filters.

**Query parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | `1` | Page number |
| `limit` | int | `20` | Items per page (max 100) |
| `search` | string | — | Case-insensitive match on name or description |
| `category_id` | UUID | — | Filter by category |
| `sort_by` | string | `created_at` | `price` or `name` |
| `sort_order` | string | `desc` | `asc` or `desc` |

**Success `200`:**
```json
{
  "success": true,
  "data": {
    "items": [ { "id": "...", "name": "...", "price": 49.99, "stock": 10, "category": { "id": "...", "name": "Electronics" }, ... } ],
    "total": 42,
    "page": 1,
    "limit": 20,
    "total_pages": 3
  }
}
```

---

#### `GET /products/{id}` — public

Single product with category details. `404` if not found.

---

#### `POST /products` — admin

```json
{
  "name": "USB-C Hub",
  "description": "7-in-1 hub",
  "price": 39.99,
  "stock": 100,
  "category_id": "<uuid>",
  "image_url": "https://example.com/hub.jpg"
}
```

`category_id` is optional (nullable).

---

#### `PUT /products/{id}` — admin

Partial update — only provided (non-zero) fields are changed. Price and stock of `0` are ignored (treated as "not provided"). To set stock to 0, use the specific field only.

---

#### `DELETE /products/{id}` — admin

Hard delete. Will fail if referenced by a non-cancelled order (`ON DELETE RESTRICT` on `order_items`).

---

### Cart

Cart is created automatically on first access. All cart endpoints require authentication.

#### `GET /cart` — user

Returns the full cart with items, product details, subtotals, and a total.

```json
{
  "success": true,
  "data": {
    "id": "...",
    "user_id": "...",
    "items": [
      {
        "id": "...",
        "product_id": "...",
        "product": { "id": "...", "name": "USB-C Hub", "price": 39.99, "stock": 99 },
        "quantity": 2,
        "subtotal": 79.98
      }
    ],
    "total": 79.98,
    "item_count": 1
  }
}
```

---

#### `POST /cart/items` — user

Add a product to the cart. If the product is already in the cart, the quantity is **added** to the existing quantity (upsert).

```json
{ "product_id": "<uuid>", "quantity": 2 }
```

Validates that the product exists and has sufficient stock. Returns the full updated cart.

**Errors:** `404` product not found, `400` insufficient stock.

---

#### `PUT /cart/items/{id}` — user

Set an item's quantity to an exact value (replaces, does not add).

```json
{ "quantity": 5 }
```

Returns the full updated cart.

---

#### `DELETE /cart/items/{id}` — user

Removes one item from the cart. Returns the full updated cart.

---

#### `DELETE /cart` — user

Removes all items from the cart.

---

### Orders

#### `POST /orders` — user

Checkout: converts the current cart into an order.

- Validates cart is not empty
- Creates the order record
- Creates order items (price is **snapshotted** from current product price)
- Atomically decrements stock for each product (fails if any item is out of stock)
- Clears the cart

All steps run in a single database transaction — either all succeed or none do.

```json
{
  "shipping_address": "123 Main St, Springfield, IL 62701",
  "notes": "Leave at the door"
}
```

**Success `201`:**
```json
{
  "success": true,
  "data": {
    "id": "...",
    "user_id": "...",
    "status": "pending",
    "total_amount": 79.98,
    "shipping_address": "123 Main St...",
    "notes": "Leave at the door",
    "created_at": "..."
  }
}
```

**Errors:** `400` empty cart, `400` insufficient stock for any item.

---

#### `GET /orders` — user/admin

Paginated order list. Users see only their own orders; admins see all orders.

**Query params:** `page`, `limit`.

---

#### `GET /orders/{id}` — user/admin

Full order with all items and product names. Users can only fetch their own orders (returns `404` for others).

---

#### `POST /orders/{id}/cancel` — user/admin

Cancel a **pending** order. Stock is restored atomically in a transaction. Users can only cancel their own orders.

**Errors:** `404` not found, `400` order is not in `pending` status.

---

#### `PUT /orders/{id}/status` — admin only

Directly set any order status.

```json
{ "status": "confirmed" }
```

Valid statuses: `pending`, `confirmed`, `shipped`, `delivered`, `cancelled`.

Note: this endpoint does **not** restore stock when setting `cancelled`. Use the cancel endpoint for that.

---

### Reviews

#### `GET /products/{id}/reviews` — public

Paginated reviews for a product, newest first. Each review includes the reviewer's name.

**Query params:** `page`, `limit`.

---

#### `POST /products/{id}/reviews` — user

One review per user per product (enforced by a `UNIQUE` constraint).

```json
{ "rating": 5, "comment": "Excellent quality!" }
```

`rating` must be 1–5. `comment` is optional.

**Errors:** `404` product not found, `409` already reviewed.

---

#### `PUT /reviews/{id}` — user/admin

Update rating and/or comment. Only the review's author or an admin can update.

```json
{ "rating": 4, "comment": "Updated opinion after more use" }
```

---

#### `DELETE /reviews/{id}` — user/admin

Delete a review. Only the review's author or an admin can delete.

---

## Authentication Flow

```
1. Register or Login
   POST /auth/register  or  POST /auth/login
   → { access_token, refresh_token, user }

2. Store both tokens client-side.

3. Call protected endpoints with:
   Authorization: Bearer <access_token>

4. When the access token expires (default 24h), use the refresh token:
   POST /auth/refresh
   Body: { "refresh_token": "<refresh_token>" }
   → { access_token }   (new access token, same refresh token)

5. When the refresh token expires (default 7d), the user must log in again.
```

The access token and refresh token are **signed with different secrets** so a leaked refresh token cannot be used to forge an access token.

---

## Role System

| Role | Set by | Capabilities |
|------|--------|--------------|
| `customer` | Default on register | Browse products, manage own cart, place/cancel own orders, write/edit/delete own reviews, update own profile |
| `admin` | Manually set in DB or via seed | Everything above + manage categories, manage products, view all orders, update any order status, delete any review |

To promote a user to admin directly in the DB:
```sql
UPDATE users SET role = 'admin' WHERE email = 'user@example.com';
```

---

## Key Business Rules

| Rule | Where enforced |
|------|----------------|
| Passwords must be ≥ 6 characters | `handlers/auth.go`, `handlers/user.go` |
| Emails are lowercased and trimmed | `handlers/auth.go`, `handlers/user.go` |
| One cart per user (auto-created) | `repository/cart.go` — `GetOrCreateByUserID` |
| Adding the same product twice adds to quantity | `repository/cart.go` — upsert with `ON CONFLICT` |
| Stock is decremented atomically at order time | `repository/order.go` — transaction |
| Order creation fails if stock < required | `repository/order.go` — `WHERE stock >= $1` check |
| Cart is cleared when order is placed | `repository/order.go` — same transaction |
| Only `pending` orders can be cancelled | `repository/order.go` — `Cancel()` checks status |
| Stock is restored when an order is cancelled | `repository/order.go` — transaction |
| Price in `order_items` is snapshotted at purchase time | `repository/order.go` — copies `product.Price` |
| One review per user per product | `UNIQUE(user_id, product_id)` in DB + pre-check in handler |
| Deleting a category nullifies products' `category_id` | `ON DELETE SET NULL` in schema |
| Deleting a product fails if it has non-cancelled order items | `ON DELETE RESTRICT` in schema |

---

## Response Format

Every response uses the same JSON envelope:

**Success:**
```json
{ "success": true, "data": { ... } }
```

**Success with message (no data body):**
```json
{ "success": true, "message": "cart cleared" }
```

**Error:**
```json
{ "success": false, "error": "description of what went wrong" }
```

HTTP status codes are used semantically:
- `200` OK
- `201` Created
- `400` Bad Request (validation failure)
- `401` Unauthorized (missing/invalid/expired token)
- `403` Forbidden (valid token, insufficient role)
- `404` Not Found
- `409` Conflict (duplicate email, duplicate review)
- `500` Internal Server Error (always generic message to client, real error logged server-side)

---

## Design Decisions

**Why no ORM?**
Raw SQL makes queries explicit and debuggable. There's no magic. Every query is exactly what hits the database.

**Why separate access and refresh token secrets?**
If the access token secret leaks, attackers can forge short-lived access tokens but cannot create long-lived refresh tokens (and vice versa).

**Why does `GET /cart` auto-create a cart?**
It removes the need for a separate "create cart" step from the client. The first time any cart endpoint is called, the cart is lazily initialized.

**Why snapshot the price in `order_items`?**
Product prices can change. The order must reflect what the customer paid at the time of purchase, not what the product costs later.

**Why use `pgtype.UUID` for nullable UUID columns only?**
`google/uuid.UUID` implements `sql.Scanner` and `driver.Valuer`, so pgx can use it directly for non-nullable UUID columns. For nullable columns (where pgx needs to write a Go `nil`), `pgtype.UUID` with its `Valid bool` field is used to distinguish NULL from a zero UUID.

**Why is `PUT /orders/{id}/status` separate from `POST /orders/{id}/cancel`?**
Cancellation has a side effect (restoring stock). A raw status update (admin tool) intentionally does not — it's for correcting status manually without touching inventory.
