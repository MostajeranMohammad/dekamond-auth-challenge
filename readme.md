## Getting Started

### Running Locally

```bash
$ go run cmd/auth/main.go
```

### Running with Docker

```bash
docker-compose up --build
```

### Example API Requests

#### Request OTP

```http
POST /api/v1/auth/request-otp
Content-Type: application/json

{
    "phone_number": "+1234567890"
}
```

#### Verify OTP & Login/Register

```http
POST /api/v1/auth/verify-otp
Content-Type: application/json

{
    "phone_number": "+1234567890",
    "otp": "123456"
}
```

#### Get User Details

```http
GET /api/v1/users/{user_id}
```

#### List Users (with pagination & search)

```http
GET /api/v1/users?search=+1234567890&page=1&limit=10
```

---

## Database Choice Justification

[Explain your choice here: e.g., PostgreSQL for reliability and scalability, Redis for fast in-memory operations, or in-memory for simplicity.]

---

## API Documentation

- Swagger UI available at `/swagger/index.html` (when running).
