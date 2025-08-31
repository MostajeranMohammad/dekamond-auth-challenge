# Usecase Unit Tests

This directory contains comprehensive unit tests for all usecases defined in the authentication challenge project. The tests are written using Go's standard testing package along with testify for assertions and gomock for mocking dependencies.

## Test Coverage

The unit tests provide comprehensive coverage of all usecase interfaces:

- **AuthService**: Tests for OTP login flow, token validation, and user creation/authentication
- **JwtUsecase**: Tests for JWT token generation, validation, and refresh token functionality
- **OtpUsecase**: Tests for OTP generation, saving, verification, and rate limiting
- **UsersService**: Tests for user retrieval and pagination functionality

## Test Structure

### AuthService Tests (`auth_test.go`)

Tests cover the complete authentication flow:

- **LoginRequestOtp**:

  - Valid phone number OTP requests
  - Invalid phone number formats
  - OTP generation failures
  - OTP save failures
  - SMS sending failures

- **VerifyLoginOTP**:

  - Successful verification for existing users
  - New user creation on first verification
  - Invalid phone numbers and OTP formats
  - OTP verification failures
  - JWT generation failures

- **ValidateToken**:

  - Valid token validation
  - Bearer prefix handling (both "Bearer" and "bearer")
  - Token with extra spaces
  - Empty and invalid tokens
  - Database errors during user retrieval

- **Integration Tests**: End-to-end flow testing

### JwtUsecase Tests (`jwt_test.go`)

Tests cover JWT token lifecycle:

- **GenerateToken**:

  - Successful token generation
  - Different user IDs
  - Empty secret key handling

- **ValidateToken**:

  - Valid token parsing
  - Invalid token formats
  - Expired tokens
  - Refresh token rejection
  - Wrong secret key validation

- **GenerateRefreshToken**:
  - Refresh token generation
  - Proper expiration times
  - Refresh token format validation

### OtpUsecase Tests (`otp_test.go`)

Tests cover OTP functionality:

- **GenerateOTP**:

  - Valid 5-digit OTP generation
  - Uniqueness testing
  - Format validation (numeric only)

- **Integration Tests**:
  - Complete OTP flow with real Redis (skipped if Redis unavailable)
  - Rate limiting tests
  - OTP consumption testing

### UsersService Tests (`users_test.go`)

Tests cover user management:

- **GetUser**:

  - Successful user retrieval
  - User not found scenarios
  - Database errors

- **GetAllUsers**:

  - Pagination logic (default and custom)
  - Phone number search filtering
  - Date range filtering
  - Repository errors
  - Empty results

- **Pagination Tests**: Edge cases for pagination calculation

## Running Tests

### Run All Tests

```bash
go test ./internal/usecases/... -v
```

### Run Tests with Coverage

```bash
go test ./internal/usecases/... -cover
```

### Run Specific Test File

```bash
go test ./internal/usecases/auth_test.go ./internal/usecases/auth.go -v
```

### Run Integration Tests (requires Redis)

```bash
go test ./internal/usecases/... -v -tags=integration
```

## Test Dependencies

The tests use the following packages:

- `github.com/stretchr/testify`: For test assertions and mocking
- `go.uber.org/mock/gomock`: For generating and using mocks
- `github.com/redis/go-redis/v9`: For Redis integration tests

## Mock Generation

Mocks are generated using gomock and stored in:

- `mockusecases/mocks.go`: Mocks for usecase interfaces
- `mockrepositories/mocks.go`: Mocks for repository interfaces

To regenerate mocks:

```bash
make generate-usecase-mocks
make generate-repository-mocks
```

## Test Features

### Comprehensive Error Testing

- All error paths are tested with appropriate error messages
- Database failures, validation errors, and external service errors

### Edge Case Coverage

- Empty inputs, malformed data, boundary conditions
- Rate limiting scenarios
- Token expiration and refresh scenarios

### Integration Testing

- End-to-end authentication flows
- Real Redis integration (when available)
- Complete user lifecycle testing

### Mocking Strategy

- External dependencies (Redis, database) are mocked
- Interfaces are used for dependency injection
- Each component is tested in isolation

## Current Test Coverage

**68.1%** statement coverage across all usecases

### Coverage by Component:

- AuthService: High coverage of authentication flows
- JwtUsecase: Complete coverage of token operations
- OtpUsecase: Core OTP logic fully tested (Redis integration tested separately)
- UsersService: Full coverage of user retrieval and pagination

## Best Practices Implemented

1. **Isolation**: Each test is independent and can run in parallel
2. **Mocking**: External dependencies are properly mocked
3. **Assertions**: Clear, descriptive assertions with proper error messages
4. **Test Structure**: Tests are organized by functionality with descriptive names
5. **Edge Cases**: Comprehensive testing of error conditions and edge cases
6. **Integration**: Real integration tests for critical flows when possible

## Future Improvements

- Add benchmark tests for performance-critical operations
- Add more Redis integration tests with test containers
- Add property-based testing for OTP generation
- Add stress testing for rate limiting functionality
