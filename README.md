# Product REST API

A RESTful API written in Go for product management, structured according to a layered
architecture and backed by PostgreSQL. The application and its database are fully
containerised through Docker Compose.

---

## Table of Contents

- [Technology Stack](#technology-stack)
- [Architecture](#architecture)
- [Request Flow](#request-flow)
- [Project Structure](#project-structure)
- [Data Model](#data-model)
- [API Reference](#api-reference)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Database Initialisation](#database-initialisation)

---

## Technology Stack

| Component        | Technology                       |
| ---------------- | -------------------------------- |
| Language         | Go 1.25                          |
| HTTP Framework   | Gin (`github.com/gin-gonic/gin`) |
| Database         | PostgreSQL 12                    |
| Database Driver  | `github.com/lib/pq`              |
| Containerisation | Docker / Docker Compose          |

---

## Architecture

The project adopts a **layered architecture**, in which each layer holds a single,
well-defined responsibility and depends exclusively on the layer immediately beneath it.
Dependencies are supplied through constructor injection during application bootstrap in
[cmd/main.go](cmd/main.go), which keeps the layers decoupled and independently testable.

| Layer          | Package      | Responsibility                                                                |
| -------------- | ------------ | ----------------------------------------------------------------------------- |
| **Controller** | `controller` | Receives HTTP requests, validates and parses input, and serialises responses. |
| **Use Case**   | `usecase`    | Hosts the business rules and orchestrates operations over the domain entities.|
| **Repository** | `repository` | Encapsulates all data access, executing SQL statements against PostgreSQL.    |
| **Model**      | `model`      | Defines the domain entities and the standard response payload.                |
| **Database**   | `db`         | Establishes and manages the PostgreSQL connection.                            |

---

## Request Flow

Every request traverses the layers in a strictly unidirectional manner:

```
   HTTP Client
        |
        v
+------------------+
|  Gin Router      |  cmd/main.go - route registration and server bootstrap
+--------+---------+
         v
+------------------+
|  Controller      |  Input validation, binding and HTTP status resolution
+--------+---------+
         v
+------------------+
|  Use Case        |  Business rules and orchestration
+--------+---------+
         v
+------------------+
|  Repository      |  Prepared statements and query execution
+--------+---------+
         v
+------------------+
|  PostgreSQL      |  Persistent storage
+------------------+
```

The response travels back along the same path in reverse order. Errors raised at any
layer are propagated upwards and translated by the controller into the appropriate HTTP
status code.

### Illustrative Flow — `GET /product/:id`

1. The router matches the route and dispatches to `ProductController.GetProductByID`.
2. The controller extracts the `id` path parameter and asserts that it is present and
   numeric; otherwise it returns `400 Bad Request`.
3. The controller delegates to `ProductUsecase.GetProductByID`.
4. The use case forwards the call to `ProductRepository.GetProductByID`.
5. The repository executes a prepared `SELECT` statement and maps the resulting row into
   a `model.Product`. When no row is found, a nil result is returned in place of an error.
6. The controller responds with `200 OK` and the product payload, or `404 Not Found`
   when the product does not exist.

---

## Project Structure

```
.
├── cmd/
│   └── main.go                      # Application entry point and route registration
├── controller/
│   └── product_controller.go        # HTTP handlers for the product resource
├── usecase/
│   └── product_usecase.go           # Business rules for the product resource
├── repository/
│   └── product_repository.go        # Data access layer (PostgreSQL)
├── model/
│   ├── product.go                   # Product domain entity
│   └── response.go                  # Standard response payload
├── db/
│   └── conn.go                      # Database connection with retry policy
├── init.sql                         # Schema definition and seed data
├── Dockerfile                       # Application image definition
├── docker-compose.yml               # Application and database orchestration
└── go.mod
```

---

## Data Model

### `Product`

| Field   | JSON Key     | Type      | Description                       |
| ------- | ------------ | --------- | --------------------------------- |
| `ID`    | `id_product` | `uint`    | Unique identifier of the product. |
| `Name`  | `name`       | `string`  | Product name.                     |
| `Price` | `price`      | `float64` | Product price.                    |

### `Response`

Standard payload used to convey informational and error messages.

| Field     | JSON Key  | Type          | Description                 |
| --------- | --------- | ------------- | --------------------------- |
| `Status`  | `status`  | `int`         | Application status code.    |
| `Message` | `message` | `string`      | Descriptive message.        |
| `Data`    | `data`    | `interface{}` | Optional accompanying data. |

---

## API Reference

**Base URL:** `http://localhost:8000`

### `GET /ping`

Health-check endpoint used to verify that the service is running.

**Response — `200 OK`**

```json
{
  "message": "pong"
}
```

---

### `GET /products`

Retrieves the complete collection of products.

**Response — `200 OK`**

```json
[
  {
    "id_product": 1,
    "name": "Playstation",
    "price": 3.5
  },
  {
    "id_product": 2,
    "name": "Fone",
    "price": 25.5
  }
]
```

**Response — `500 Internal Server Error`** — returned when the query fails.

**Example**

```bash
curl http://localhost:8000/products
```

---

### `GET /product/:id`

Retrieves a single product by its unique identifier.

| Parameter | Location | Type      | Required | Description                       |
| --------- | -------- | --------- | -------- | --------------------------------- |
| `id`      | Path     | `integer` | Yes      | Unique identifier of the product. |

**Response — `200 OK`**

```json
{
  "id_product": 1,
  "name": "Playstation",
  "price": 3.5
}
```

**Response — `400 Bad Request`** — the identifier is absent or is not a valid number.

```json
{
  "status": 0,
  "message": "ID must be a number",
  "data": null
}
```

**Response — `404 Not Found`** — no product exists for the supplied identifier.

```json
{
  "status": 0,
  "message": "Product not found",
  "data": null
}
```

**Response — `500 Internal Server Error`** — returned when the query fails.

**Example**

```bash
curl http://localhost:8000/product/1
```

---

### `POST /product`

Creates a new product and returns the persisted representation, including the identifier
generated by the database.

**Request Body**

```json
{
  "name": "Teclado Mecanico",
  "price": 349.90
}
```

**Response — `201 Created`**

```json
{
  "id_product": 4,
  "name": "Teclado Mecanico",
  "price": 349.9
}
```

**Response — `400 Bad Request`** — the request body is malformed or cannot be bound to
the `Product` entity.

**Response — `500 Internal Server Error`** — returned when the insertion fails.

**Example**

```bash
curl -X POST http://localhost:8000/product \
  -H "Content-Type: application/json" \
  -d '{"name":"Teclado Mecanico","price":349.90}'
```

---

### Endpoint Summary

| Method | Endpoint       | Description                   | Success Status |
| ------ | -------------- | ----------------------------- | -------------- |
| `GET`  | `/ping`        | Service health check.         | `200 OK`       |
| `GET`  | `/products`    | List all products.            | `200 OK`       |
| `GET`  | `/product/:id` | Retrieve a product by its ID. | `200 OK`       |
| `POST` | `/product`     | Create a new product.         | `201 Created`  |

---

## Getting Started

### Prerequisites

- Docker and Docker Compose, **or**
- Go 1.25 or later together with a running PostgreSQL 12 instance.

### Running with Docker Compose (recommended)

```bash
docker compose up --build
```

This command builds the application image, starts the PostgreSQL container, applies the
schema and seed data defined in [init.sql](init.sql), and exposes the API on port `8000`.

To stop the environment:

```bash
docker compose down
```

To stop the environment and discard the persisted database volume:

```bash
docker compose down -v
```

### Running Locally

Ensure that a PostgreSQL instance is reachable and that the schema defined in `init.sql`
has been applied, then execute:

```bash
go mod download
go run cmd/main.go
```

The service listens on port `8000`.

---

## Configuration

The database connection is configured through environment variables. Sensible defaults
are applied whenever a variable is not provided.

| Variable      | Default     | Description         |
| ------------- | ----------- | ------------------- |
| `DB_HOST`     | `localhost` | Database host name. |
| `DB_PORT`     | `5432`      | Database port.      |
| `DB_USER`     | `postgres`  | Database user.      |
| `DB_PASSWORD` | `1234`      | Database password.  |
| `DB_NAME`     | `postgres`  | Database name.      |

> **Note:** the default credentials are intended for local development only. They should
> be replaced with securely managed secrets before any deployment to a shared or
> production environment.

### Connection Resilience

On start-up the application attempts to establish the database connection up to six
times, pausing five seconds between attempts. This retry policy accommodates the
start-up delay of the PostgreSQL container when the stack is launched through Docker
Compose.

---

## Database Initialisation

The [init.sql](init.sql) script is mounted into the PostgreSQL container initialisation
directory and is executed automatically on the first start of the database volume. It
creates the `product` table and inserts a small set of sample records.

```sql
CREATE TABLE IF NOT EXISTS product (
    id SERIAL PRIMARY KEY,
    product_name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);
```

Because the script runs only when the data directory is empty, any subsequent change to
`init.sql` requires the volume to be recreated with `docker compose down -v`.
