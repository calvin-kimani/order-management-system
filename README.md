# Order Management System

## Features
- Complete order management with customers/products
- M-Pesa payment integration
- Dockerized microservices
- Unit tests & logging

## Setup
1. Clone repository
2. Create `.env` file from `.env.example`
3. Get M-Pesa credentials
4. Start services:
```bash
docker-compose up --build
```

## 1. Orders Service
### Create Customer

```bash
curl -X POST http://localhost:8080/customers \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'
```

### Create Product

```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name": "Premium Coffee", "price": 15.99}'
```

### Create Order

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": 1,
    "products": [{"id": 1}]
  }'
```

### Verify Order

```bash
curl http://localhost:8080/orders/1
```

### Verify Order Status Update after payment
After payment simulation:

```bash
curl http://localhost:8080/orders/1/status
```
Expected Response:

```bash
{
  "id": 1,
  "status": "paid",
  "total": 15.99
}
```

## 2. Payment Service
### Initiate Payment (Sandbox Test)

```bash
curl -X POST http://localhost:8081/payments \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": 1,
    "amount": "1",
    "phone": "254708374149"  # Recipient number
  }'
```
#### Expected Response

```json
{
  "message": "Payment initiated",
  "checkout_request_id": "ws_CO_191220231510440123456789"
}
```

#### Simulate M-Pesa Callback
1. Install ngrok to expose your local service:

```bash
ngrok http 8081
```

2. Update .env with ngrok URL:

```bash
MPESA_CALLBACK_URL=https://your-ngrok-id.ngrok.io/callback
```

3. Restart services:

```bash
docker-compose restart paymentservice
```

**M-Pesa sandbox will now send callbacks to your local service through ngrok.**

## End-to-End Test Scenario

```bash
# 1. Create customer
curl -X POST http://localhost:8080/customers -d '{"name":"Test User","email":"test@example.com"}'

# 2. Create product
curl -X POST http://localhost:8080/products -d '{"name":"Test Product","price":9.99}'

# 3. Create order
curl -X POST http://localhost:8080/orders -d '{"customer_id":1,"products":[{"id":1}]}'

# 4. Process payment
curl -X POST http://localhost:8081/payments -d '{"order_id":1,"amount":"1","phone":"254708374149"}'

# 5. Verify system state
curl http://localhost:8080/orders/1/status
```

## 3. Tests

### Run all tests with coverage

```bash
go test -v -coverprofile=coverage.out ./...
```

### View coverage report
```bash
go tool cover -html=coverage.out
```