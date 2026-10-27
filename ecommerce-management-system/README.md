# Ecommerce Management System

This is an Ecommerce System using Go, GraphQL ('graphql-go'), Gin, HashiCorp Consul, and PostgreSQL.
---

## 📌 Features

* **Create Order**: Creates order with customer ID, product ID, and quantity while automatically fetching live product prices with circuit-breaking resilience (`gobreaker`).
* **Get Order**: Get details of order by its order ID.
* **List Orders**: Get all the order details.
* **Validation**:
  * Customer ID and Product ID must be positive integers.
  * Quantity must be greater than zero.
  * Total amount is automatically calculated based on product pricing.
  * Order status must be one of `CREATED`, `CONFIRMED`, or `CANCELLED`.
* **Security & Observability**:
  * JWT (HS256) authentication with role-based access control (`admin` vs `user`).
  * Layered health checks (`/health/live`, `/health/ready`) and automated `X-Request-ID` tracing logs.

---

## 📂 Project Structure

```text
ecommerce-microservices/
├── .env
├── .gitignore
├── buildspec.yml
├── Procfile
├── go.mod
├── go.sum
├── schema.sql
├── register-consul.ps1
├── examples.graphql
├── soapui.txt
├── gateway/
│   ├── main.go
│   ├── config/
│   │   └── config.go
│   ├── client/
│   │   └── consul_client.go
│   └── graphql/
│       ├── schema.go
│       └── handler.go
└── services/
    ├── auth/
    │   ├── main.go
    │   ├── auth/
    │   │   └── auth.go
    │   ├── config/
    │   │   └── config.go
    │   └── controller/
    │       └── auth_controller.go
    ├── product/
    │   ├── main.go
    │   ├── config/
    │   │   └── config.go
    │   ├── model/
    │   │   └── product.go
    │   ├── repository/
    │   │   └── product_repository.go
    │   ├── service/
    │   │   └── product_service.go
    │   ├── controller/
    │   │   └── product_controller.go
    │   ├── database/
    │   │   └── postgres.go
    │   ├── health/
    │   │   └── health.go
    │   ├── logging/
    │   │   └── logger.go
    │   └── middleware/
    │       ├── auth.go
    │       ├── logging.go
    │       ├── recovery.go
    │       └── request_id.go
    └── order/
        ├── main.go
        ├── config/
        │   └── config.go
        ├── model/
        │   └── order.go
        ├── repository/
        │   └── order_repository.go
        ├── service/
        │   └── order_service.go
        ├── controller/
        │   └── order_controller.go
        ├── client/
        │   └── product_client.go
        ├── database/
        │   └── postgres.go
        ├── health/
        │   └── health.go
        ├── logging/
        │   └── logger.go
        ├── middleware/
        │   ├── auth.go
        │   ├── logging.go
        │   ├── recovery.go
        │   └── request_id.go
        └── tests/
            └── concurrent_order_test.go
```

---

## 🚀 Getting Started

### 1. Clone the repository

```bash
git clone <repository-url>
cd ecommerce-microservices
```

### 2. Database & Service Discovery Setup

Execute the schema against PostgreSQL (`assessment5` database):
```bash
psql -U postgres -d assessment5 -f schema.sql
```

Register services with HashiCorp Consul (running on port `8500`):
```powershell
./register-consul.ps1
```

### 3. Run the application

Run each microservice and the API Gateway in separate terminal windows:

```bash
# 1. Run Auth Service (Port 8083)
go run ./services/auth

# 2. Run Product Service (Port 8081)
go run ./services/product

# 3. Run Order Service (Port 8082)
go run ./services/order

# 4. Run API Gateway (Port 8080)
go run ./gateway
```

---

## 💻 Usage Example

### 1. Obtain a JWT Bearer Token

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'
```

### 2. REST API Example (via Gateway with Token)

```bash
# Create Order
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"customerId": 101, "productId": 1, "quantity": 2}'

# Get Order by ID
curl -X GET http://localhost:8080/api/v1/orders/1 \
  -H "Authorization: Bearer <TOKEN>"
```

### 3. GraphQL Example

Open the interactive playground in your browser at `http://localhost:8080/playground` or send a query to `http://localhost:8080/graphql`:

```graphql
# Query single order
query GetOrderByID {
  order(id: 1) {
    id
    customerId
    productId
    quantity
    total
    status
  }
}

# Create Order Mutation
mutation CreateOrder {
  createOrder(customerId: 101, productId: 1, quantity: 2, total: 179.98) {
    id
    customerId
    productId
    quantity
    total
    status
  }
}
```

### 4. Concurrency Performance Test

Verify 100 concurrent requests with the race detector:
```bash
go test -v -race ./services/order/tests/...
```