# Dekamond Auth Challenge

## Overview

This project implements OTP-based login and registration in Golang, featuring rate limiting, user management, RESTful APIs, and containerization.

---

## Features

### 1. OTP Login & Registration

- Users send their phone number; system generates a random OTP.
- OTP is printed to the console (no SMS sending).
- OTP is stored temporarily (DB or in-memory) and expires after 2 minutes.
- User submits phone number + OTP:
  - If OTP is valid & not expired:
    - Register new user if not existing.
    - Log in existing user otherwise.
- Upon success, a JWT token is returned.

### 2. Rate Limiting

- Limit OTP requests to **max 3 per phone number within 10 minutes**.

### 3. User Management

- REST endpoints to:
  - Retrieve single user details.
  - Retrieve list of users with pagination and search (by phone number or other fields).
- Store at minimum:
  - Phone number
  - Registration date

### 4. Database

- **Choice:** [Specify your DB here, e.g., PostgreSQL, Redis, or in-memory]
- If using a DB:
  - Set up with `docker-compose`.
- If not using a DB:
  - Use in-memory storage for simplicity.

### 5. API Documentation

- All operations exposed via REST APIs.
- Documented with Swagger/OpenAPI.

### 6. Architecture & Best Practices

- Clean, maintainable architecture.
- Clear separation of responsibilities in code.

### 7. Containerization

- Application is Dockerized.
- DB included in `docker-compose` (if applicable).

---

## Deliverables

- Source code.
- Documentation:
  - How to run locally.
  - How to run with Docker.
  - Example API requests & responses.
  - Database choice justification.

---

## Getting Started

### Running Locally

```bash
go run main.go
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

---

## Timeframe

**Complete within 48 hours.**
