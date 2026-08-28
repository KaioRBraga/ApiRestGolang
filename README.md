<div align="center">

# Product REST API

**A RESTful API written in Go for product management, built on a layered architecture and backed by PostgreSQL.**

<br>

<img src="https://skillicons.dev/icons?i=go,postgres,docker,git&theme=light" alt="Go, PostgreSQL, Docker, Git" height="60">

<br><br>

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
[![Gin](https://img.shields.io/badge/Gin-1.12-008ECF?style=for-the-badge&logo=gin&logoColor=white)](https://gin-gonic.com)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-12-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com)

![Architecture](https://img.shields.io/badge/architecture-layered-6E56CF?style=flat-square)
![API](https://img.shields.io/badge/API-REST-FF6B35?style=flat-square)
![Port](https://img.shields.io/badge/port-8000-555555?style=flat-square)
![Status](https://img.shields.io/badge/status-active-3FB950?style=flat-square)

</div>

---

## Table of Contents

| Section | Description |
| :--- | :--- |
| [Technology Stack](#technology-stack) | Languages, frameworks and infrastructure. |
| [Architecture](#architecture) | Layer responsibilities and dependency direction. |
| [Request Flow](#request-flow) | How a request traverses the application. |
| [Project Structure](#project-structure) | Directory and file organisation. |
| [Data Model](#data-model) | Domain entities and database schema. |
| [API Reference](#api-reference) | Endpoints, payloads and status codes. |
| [Getting Started](#getting-started) | Installation and execution. |
| [Configuration](#configuration) | Environment variables and defaults. |
| [Database Initialisation](#database-initialisation) | Schema bootstrap and seed data. |

---

## Technology Stack

<div align="center">

| | Component | Technology | Version |
| :---: | :--- | :--- | :--- |
| <img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/go/go-original-wordmark.svg" width="34"> | **Language** | Go | `1.25` |
| <img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/go/go-original.svg" width="34"> | **HTTP Framework** | Gin (`gin-gonic/gin`) | `1.12.0` |
| <img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/postgresql/postgresql-original.svg" width="34"> | **Database** | PostgreSQL | `12` |
| <img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/postgresql/postgresql-plain.svg" width="34"> | **Driver** | `lib/pq` | `1.12.3` |
| <img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/docker/docker-original.svg" width="34"> | **Containerisation** | Docker / Docker Compose | — |

</div>

---

## Architecture

The project adopts a **layered architecture**, in which each layer holds a single, well-defined
responsibility and depends exclusively on the layer immediately beneath it. Dependencies are
supplied through constructor injection during application bootstrap in [cmd/main.go](cmd/main.go),
which keeps the layers decoupled and independently testable.

```mermaid
flowchart TD
    Client(["HTTP Client"])

    subgraph APP ["Application — Go 1.25 · Gin"]
        direction TB
        Router["<b>Router</b><br/><code>cmd/main.go</code><br/><i>Route registration · Bootstrap</i>"]
        Controller["<b>Controller Layer</b><br/><code>controller/</code><br/><i>Binding · Validation · HTTP status</i>"]
        Usecase["<b>Use Case Layer</b><br/><code>usecase/</code><br/><i>Business rules · Orchestration</i>"]
        Repository["<b>Repository Layer</b><br/><code>repository/</code><br/><i>Prepared statements · Row mapping</i>"]
    end

    Model["<b>Model</b><br/><code>model/</code><br/><i>Product · Response</i>"]
    Database[("<b>PostgreSQL 12</b><br/><i>table: product</i>")]

    Client -- "HTTP request" --> Router
    Router --> Controller
    Controller --> Usecase
    Usecase --> Repository
    Repository -- "SQL" --> Database
    Database -. "rows" .-> Repository
    Repository -. "entities" .-> Usecase
    Usecase -. "entities" .-> Controller
    Controller -. "JSON response" .-> Client

    Controller -.-> Model
    Usecase -.-> Model
    Repository -.-> Model

    classDef client fill:#FF6B35,stroke:#B23F14,stroke-width:2px,color:#FFFFFF
    classDef layer fill:#E8F4FD,stroke:#00ADD8,stroke-width:2px,color:#0B3954
    classDef model fill:#F3EFFF,stroke:#6E56CF,stroke-width:2px,color:#2A1B57
    classDef data fill:#E4ECFB,stroke:#4169E1,stroke-width:2px,color:#0B2559

    class Client client
    class Router,Controller,Usecase,Repository layer
    class Model model
    class Database data
```

### Layer Responsibilities

| Layer | Package | Responsibility |
| :--- | :--- | :--- |
| **Controller** | `controller` | Receives HTTP requests, validates and parses input, and serialises responses. |
| **Use Case** | `usecase` | Hosts the business rules and orchestrates operations over the domain entities. |
| **Repository** | `repository` | Encapsulates all data access, executing SQL statements against PostgreSQL. |
| **Model** | `model` | Defines the domain entities and the standard response payload. |
| **Database** | `db` | Establishes and manages the PostgreSQL connection. |

---

## Request Flow

Every request traverses the layers in a strictly unidirectional manner, and the response
travels back along the same path in reverse order. Errors raised at any layer are propagated
upwards and translated by the controller into the appropriate HTTP status code.

### Retrieval — `GET /product/:id`

```mermaid
sequenceDiagram
    autonumber
    actor C as Client
    participant R as Gin Router
    participant CT as Controller
    participant UC as Use Case
    participant RP as Repository
    participant DB as PostgreSQL

    C->>R: GET /product/1
    R->>CT: GetProductByID(ctx)

    alt id absent or non-numeric
        CT-->>C: 400 Bad Request<br/>{"message":"ID must be a number"}
    else id valid
        CT->>UC: GetProductByID(1)
        UC->>RP: GetProductByID(1)
        RP->>DB: SELECT * FROM product WHERE id = $1
        DB-->>RP: row / sql.ErrNoRows
        RP-->>UC: *model.Product / nil
        UC-->>CT: *model.Product / nil

        alt product found
            CT-->>C: 200 OK<br/>{"id_product":1,"name":"...","price":...}
        else product not found
            CT-->>C: 404 Not Found<br/>{"message":"Product not found"}
        end
    end
```

### Creation — `POST /product`

```mermaid
sequenceDiagram
    autonumber
    actor C as Client
    participant R as Gin Router
    participant CT as Controller
    participant UC as Use Case
    participant RP as Repository
    participant DB as PostgreSQL

    C->>R: POST /product<br/>{"name":"Teclado","price":349.90}
    R->>CT: CreateProduct(ctx)
    CT->>CT: ShouldBindJSON(&product)

    alt malformed body
        CT-->>C: 400 Bad Request
    else body valid
        CT->>UC: CreateProduct(product)
        UC->>RP: CreateProduct(product)
        RP->>DB: INSERT INTO product (product_name, price)<br/>VALUES ($1, $2) RETURNING id
        DB-->>RP: generated id
        RP-->>UC: id
        UC->>UC: product.ID = id
        UC-->>CT: model.Product
        CT-->>C: 201 Created<br/>{"id_product":4,"name":"Teclado","price":349.9}
    end
```

---

## Project Structure

```
restAPI/
│
├── cmd/
│   └── main.go                    # Entry point · route registration · DI bootstrap
│
├── controller/
│   └── product_controller.go      # HTTP handlers for the product resource
│
├── usecase/
│   └── product_usecase.go         # Business rules for the product resource
│
├── repository/
│   └── product_repository.go      # Data access layer (PostgreSQL)
│
├── model/
│   ├── product.go                 # Product domain entity
│   └── response.go                # Standard response payload
│
├── db/
│   └── conn.go                    # Database connection with retry policy
│
├── init.sql                       # Schema definition and seed data
├── Dockerfile                     # Application image definition
├── docker-compose.yml             # Application and database orchestration
└── go.mod
```

---

## Data Model

```mermaid
erDiagram
    PRODUCT {
        SERIAL id PK "Auto-generated identifier"
        VARCHAR product_name "NOT NULL, max 255 chars"
        DECIMAL price "NOT NULL, precision (10,2)"
    }
```

### `Product` — domain entity

| Field | JSON Key | Go Type | SQL Column | Description |
| :--- | :--- | :--- | :--- | :--- |
| `ID` | `id_product` | `uint` | `id` | Unique identifier of the product. |
| `Name` | `name` | `string` | `product_name` | Product name. |
| `Price` | `price` | `float64` | `price` | Product price. |

> [!IMPORTANT]
> The JSON key for the identifier is **`id_product`**, not `id`.

### `Response` — standard message payload

| Field | JSON Key | Go Type | Description |
| :--- | :--- | :--- | :--- |
| `Status` | `status` | `int` | Application status code. |
| `Message` | `message` | `string` | Descriptive message. |
| `Data` | `data` | `interface{}` | Optional accompanying data. |

---

## API Reference

<div align="center">

**Base URL** · `http://localhost:8000`

</div>

### Endpoint Summary

| | Endpoint | Description | Success |
| :--- | :--- | :--- | :--- |
| ![GET](https://img.shields.io/badge/GET-61AFFE?style=for-the-badge&logoColor=white) | `/ping` | Service health check. | `200 OK` |
| ![GET](https://img.shields.io/badge/GET-61AFFE?style=for-the-badge&logoColor=white) | `/products` | List all products. | `200 OK` |
| ![GET](https://img.shields.io/badge/GET-61AFFE?style=for-the-badge&logoColor=white) | `/product/:id` | Retrieve a product by its ID. | `200 OK` |
| ![POST](https://img.shields.io/badge/POST-49CC90?style=for-the-badge&logoColor=white) | `/product` | Create a new product. | `201 Created` |

---

<details open>
<summary><h3>&nbsp;<img src="https://img.shields.io/badge/GET-61AFFE?style=flat-square" align="center">&nbsp; <code>/ping</code> — Health check</h3></summary>

Verifies that the service is running.

**Response** · `200 OK`

```json
{
  "message": "pong"
}
```

**Example**

```bash
curl http://localhost:8000/ping
```

</details>

---

<details open>
<summary><h3>&nbsp;<img src="https://img.shields.io/badge/GET-61AFFE?style=flat-square" align="center">&nbsp; <code>/products</code> — List products</h3></summary>

Retrieves the complete collection of products.

**Response** · `200 OK`

```json
[
  { "id_product": 1, "name": "Playstation", "price": 3.5 },
  { "id_product": 2, "name": "Fone",        "price": 25.5 },
  { "id_product": 3, "name": "Carregador",  "price": 30.99 }
]
```

| Status | Condition |
| :--- | :--- |
| ![200](https://img.shields.io/badge/200-3FB950?style=flat-square) `OK` | Collection returned successfully. |
| ![500](https://img.shields.io/badge/500-F93E3E?style=flat-square) `Internal Server Error` | The query failed. |

**Example**

```bash
curl http://localhost:8000/products
```

</details>

---

<details open>
<summary><h3>&nbsp;<img src="https://img.shields.io/badge/GET-61AFFE?style=flat-square" align="center">&nbsp; <code>/product/:id</code> — Retrieve by ID</h3></summary>

Retrieves a single product by its unique identifier.

**Path parameters**

| Parameter | Type | Required | Description |
| :--- | :--- | :---: | :--- |
| `id` | `integer` | Yes | Unique identifier of the product. |

**Response** · `200 OK`

```json
{
  "id_product": 1,
  "name": "Playstation",
  "price": 3.5
}
```

| Status | Condition | Body |
| :--- | :--- | :--- |
| ![200](https://img.shields.io/badge/200-3FB950?style=flat-square) `OK` | Product found. | `Product` |
| ![400](https://img.shields.io/badge/400-FCA130?style=flat-square) `Bad Request` | Identifier absent or non-numeric. | `{"message":"ID must be a number"}` |
| ![404](https://img.shields.io/badge/404-FCA130?style=flat-square) `Not Found` | No product for the supplied identifier. | `{"message":"Product not found"}` |
| ![500](https://img.shields.io/badge/500-F93E3E?style=flat-square) `Internal Server Error` | The query failed. | error |

**Example**

```bash
curl http://localhost:8000/product/1
```

</details>

---

<details open>
<summary><h3>&nbsp;<img src="https://img.shields.io/badge/POST-49CC90?style=flat-square" align="center">&nbsp; <code>/product</code> — Create product</h3></summary>

Creates a new product and returns the persisted representation, including the identifier
generated by the database.

**Request body**

```json
{
  "name": "Teclado Mecanico",
  "price": 349.90
}
```

**Response** · `201 Created`

```json
{
  "id_product": 4,
  "name": "Teclado Mecanico",
  "price": 349.9
}
```

| Status | Condition |
| :--- | :--- |
| ![201](https://img.shields.io/badge/201-3FB950?style=flat-square) `Created` | Product persisted successfully. |
| ![400](https://img.shields.io/badge/400-FCA130?style=flat-square) `Bad Request` | Body malformed or not bindable to `Product`. |
| ![500](https://img.shields.io/badge/500-F93E3E?style=flat-square) `Internal Server Error` | The insertion failed. |

**Example**

```bash
curl -X POST http://localhost:8000/product \
  -H "Content-Type: application/json" \
  -d '{"name":"Teclado Mecanico","price":349.90}'
```

</details>

---

## Getting Started

### Prerequisites

<div align="center">

| Option | Requirements |
| :--- | :--- |
| ![Docker](https://img.shields.io/badge/Recommended-2496ED?style=flat-square&logo=docker&logoColor=white) | Docker and Docker Compose |
| ![Go](https://img.shields.io/badge/Local-00ADD8?style=flat-square&logo=go&logoColor=white) | Go 1.25+ and a running PostgreSQL 12 instance |

</div>

### Running with Docker Compose

```bash
docker compose up --build
```

This command builds the application image, starts the PostgreSQL container, applies the schema
and seed data defined in [init.sql](init.sql), and exposes the API on port `8000`.

```mermaid
flowchart LR
    subgraph HOST ["Host machine"]
        direction LR
        subgraph NET ["Docker Compose network"]
            direction LR
            APP["<b>go_app</b><br/>image: rest-api<br/>:8000"]
            DB[("<b>go_db</b><br/>postgres:12<br/>:5432")]
            VOL[("<b>db_data</b><br/><i>named volume</i>")]
        end
    end

    User(["Client"]) -- "localhost:8000" --> APP
    APP -- "DB_HOST=go_db" --> DB
    DB --- VOL
    INIT["init.sql"] -. "mounted into<br/>docker-entrypoint-initdb.d" .-> DB

    classDef app fill:#E8F4FD,stroke:#00ADD8,stroke-width:2px,color:#0B3954
    classDef db fill:#E4ECFB,stroke:#4169E1,stroke-width:2px,color:#0B2559
    classDef vol fill:#F3EFFF,stroke:#6E56CF,stroke-width:2px,color:#2A1B57
    classDef user fill:#FF6B35,stroke:#B23F14,stroke-width:2px,color:#FFFFFF
    classDef file fill:#FFF6E0,stroke:#FCA130,stroke-width:2px,color:#5A3A00

    class APP app
    class DB db
    class VOL vol
    class User user
    class INIT file
```

**Lifecycle commands**

| Command | Effect |
| :--- | :--- |
| `docker compose up --build` | Build the image and start the full stack. |
| `docker compose down` | Stop the environment, preserving the database volume. |
| `docker compose down -v` | Stop the environment and **discard** the database volume. |
| `docker compose logs -f go_app` | Follow the application logs. |

### Running Locally

Ensure that a PostgreSQL instance is reachable and that the schema defined in `init.sql` has
been applied, then execute:

```bash
go mod download
go run cmd/main.go
```

The service listens on port `8000`.

---

## Configuration

The database connection is configured through environment variables. Sensible defaults are
applied whenever a variable is not provided.

| Variable | Default | Description |
| :--- | :--- | :--- |
| `DB_HOST` | `localhost` | Database host name. |
| `DB_PORT` | `5432` | Database port. |
| `DB_USER` | `postgres` | Database user. |
| `DB_PASSWORD` | `1234` | Database password. |
| `DB_NAME` | `postgres` | Database name. |

> [!WARNING]
> The default credentials are intended for local development only. They should be replaced with
> securely managed secrets before any deployment to a shared or production environment.

### Connection Resilience

On start-up the application attempts to establish the database connection up to **six times**,
pausing **five seconds** between attempts. This retry policy accommodates the start-up delay of
the PostgreSQL container when the stack is launched through Docker Compose.

```mermaid
flowchart LR
    Start(["Start"]) --> Open["sql.Open"]
    Open --> Ping{"db.Ping()<br/>succeeded?"}
    Ping -- "yes" --> Ok(["Connected"])
    Ping -- "no" --> Check{"attempts<br/>&lt; 6 ?"}
    Check -- "yes" --> Wait["sleep 5s"]
    Wait --> Open
    Check -- "no" --> Fail(["panic"])

    classDef ok fill:#E6F6EA,stroke:#3FB950,stroke-width:2px,color:#04331A
    classDef fail fill:#FDECEC,stroke:#F93E3E,stroke-width:2px,color:#4A0A0A
    classDef step fill:#E8F4FD,stroke:#00ADD8,stroke-width:2px,color:#0B3954
    classDef dec fill:#FFF6E0,stroke:#FCA130,stroke-width:2px,color:#5A3A00

    class Ok ok
    class Fail fail
    class Start,Open,Wait step
    class Ping,Check dec
```

---

## Database Initialisation

The [init.sql](init.sql) script is mounted into the PostgreSQL container initialisation
directory and is executed automatically on the first start of the database volume. It creates
the `product` table and inserts a small set of sample records.

```sql
CREATE TABLE IF NOT EXISTS product (
    id SERIAL PRIMARY KEY,
    product_name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);
```

**Seed data**

| `id` | `product_name` | `price` |
| :---: | :--- | ---: |
| 1 | Playstation | 3.50 |
| 2 | Fone | 25.50 |
| 3 | Carregador | 30.99 |

> [!NOTE]
> The script runs **only when the data directory is empty**. Any subsequent change to
> `init.sql` therefore requires the volume to be recreated with `docker compose down -v`.

---

<div align="center">

<sub>Built with <img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/go/go-original.svg" width="16" align="center"> Go · <img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/postgresql/postgresql-original.svg" width="14" align="center"> PostgreSQL · <img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/docker/docker-original.svg" width="16" align="center"> Docker</sub>

</div>
