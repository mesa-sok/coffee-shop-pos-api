# Coffee Shop POS API

A modern Coffee Shop Point of Sale (POS) backend built with Go, Gin, and PostgreSQL, following Clean Architecture principles.

## Tech Stack

- **Language:** Go (Golang)
- **Framework:** [Gin Gonic](https://github.com/gin-gonic/gin)
- **Database:** PostgreSQL
- **Database Library:** [sqlx](https://github.com/jmoiron/sqlx)
- **Containerization:** Docker & Docker Compose

## Architecture

This project follows **Clean Architecture** to ensure separation of concerns, testability, and maintainability.

- **Domain:** Contains business entities and repository/usecase interfaces.
- **Usecase:** Contains business logic.
- **Repository:** Handles data persistence (PostgreSQL).
- **Delivery (HTTP):** Handles external communication via REST API (Gin).
- **Configs:** Configuration management using environment variables.

## Project Structure

```text
├── cmd
│   └── api            # Application entry point
├── configs            # Configuration loading
├── internal
│   ├── delivery       # HTTP handlers and routing
│   ├── domain         # Business entities and interfaces
│   ├── repository     # Data access layer
│   └── usecase        # Business logic layer
├── migrations         # SQL migration files
└── docker-compose.yml # Docker configuration
```

## Getting Started

### Prerequisites

- Go 1.23+
- Docker & Docker Compose

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd coffee-shop-pos
   ```

2. Copy the environment file:
   ```bash
   cp .env.example .env
   ```

3. Start the database using Docker:
   ```bash
   docker-compose up -d
   ```

4. Install Go dependencies:
   ```bash
   go mod download
   ```

### Running the Application

To start the API server:

```bash
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`.

## API Endpoints

### Menu Management

| Method | Endpoint             | Description             |
|--------|----------------------|-------------------------|
| POST   | `/api/v1/menu`       | Create a new menu item  |
| GET    | `/api/v1/menu`       | Fetch all menu items    |
| GET    | `/api/v1/menu/:id`   | Get a menu item by ID   |
| PUT    | `/api/v1/menu/:id`   | Update a menu item      |
| DELETE | `/api/v1/menu/:id`   | Delete a menu item      |

### Example JSON Body for Create/Update

```json
{
  "name": "Cappuccino",
  "description": "Espresso with steamed milk foam",
  "price": 4.50,
  "category": "Coffee",
  "is_available": true
}
```

## License

MIT
