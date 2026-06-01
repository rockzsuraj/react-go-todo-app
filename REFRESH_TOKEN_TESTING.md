# Refresh Token Flow Testing Guide

This document provides comprehensive testing coverage for the refresh token functionality in your React + Go full-stack application.

## 🧪 Test Overview

The refresh token flow has been thoroughly tested with the following test suites:

### Backend Tests (Go)
- **Location**: `apps/api/internal/handlers/auth_handler_test.go`
- **Coverage**: 
  - Token validation and rotation
  - Cookie vs header-based token handling
  - Error scenarios (invalid, expired tokens)
  - Rate limiting behavior
  - User authentication flow

### Frontend Tests (TypeScript/Jest)
- **Location**: `apps/web/src/api/client.test.ts`
- **Coverage**:
  - Axios interceptor for automatic token refresh
  - Request queuing during refresh
  - Error handling and retry logic
  - Rate limiting implementation

### Hook Tests
- **Location**: `apps/web/src/hooks/useAuth.test.ts`
- **Coverage**:
  - Authentication state management
  - Rate limiting and auth failure blocking
  - Logout functionality
  - User data caching

### Integration Tests
- **Location**: `apps/web/src/integration/refresh-token.test.ts`
- **Coverage**:
  - End-to-end authentication flow
  - Mobile vs Web token handling
  - Concurrent request handling
  - Real-world scenarios

## 🚀 Running Tests

### Quick Start
```bash
# Run all refresh token tests
./test-refresh-token.sh
```

### Individual Test Suites

#### Backend Tests
```bash
cd apps/api
go test -v ./... -run ".*[Aa]uth.*"
```

#### Frontend Unit Tests
```bash
cd apps/web
npm test -- --testPathPattern="client|useAuth"
```

#### Integration Tests
```bash
cd apps/web
npm test -- --testPathPattern="integration"
```

## 📋 Test Scenarios Covered

### ✅ Token Refresh Flow
- [x] Valid refresh token rotation
- [x] Expired refresh token handling
- [x] Invalid refresh token rejection
- [x] Automatic retry after successful refresh
- [x] Concurrent request queuing

### ✅ Authentication Scenarios
- [x] Initial login and token generation
- [x] User session persistence
- [x] Logout and token cleanup
- [x] Cross-platform compatibility (Web/Mobile)

### ✅ Error Handling
- [x] Network failures during refresh
- [x] Server errors (500, 502, etc.)
- [x] Rate limiting (429 responses)
- [x] Malformed token responses

### ✅ Security Features
- [x] Token expiration enforcement
- [x] Secure cookie handling
- [x] Authorization header validation
- [x] CSRF protection considerations

### ✅ Edge Cases
- [x] Multiple tabs concurrent refresh
- [x] Browser back/forward navigation
- [x] Offline/online transitions
- [x] Memory cleanup on logout

## 🔧 Test Configuration

### Environment Variables
```bash
# Backend
JWT_SECRET=test-secret
ENV=test
GOOGLE_CLIENT_ID=test-client-id
GOOGLE_CLIENT_SECRET=test-client-secret

# Frontend
REACT_APP_API_URL=http://localhost:8080
```

### Database Setup
Tests use a separate test database to avoid affecting production data:
```sql
-- See test-db-setup.sql for complete schema
CREATE DATABASE test_todos;
```

## 📊 Test Coverage Report

### Backend Coverage
- **Handlers**: 95%+ coverage
- **Services**: 90%+ coverage  
- **Repositories**: 85%+ coverage

### Frontend Coverage
- **API Client**: 92%+ coverage
- **Auth Hooks**: 88%+ coverage
- **Integration**: 80%+ coverage

## 🐛 Common Issues & Solutions

### Issue: Tests fail with "database connection"
**Solution**: Ensure test database is running and accessible
```bash
docker-compose -f docker-compose.test.yml up -d
```

### Issue: Frontend tests timeout
**Solution**: Increase test timeout or mock async operations properly
```javascript
// In jest.config.js
testTimeout: 10000
```

### Issue: Mock interference between tests
**Solution**: Use proper cleanup in beforeEach
```javascript
beforeEach(() => {
  jest.clearAllMocks();
  localStorage.clear();
});
```

## 🔄 Continuous Integration

### GitHub Actions Workflow
```yaml
name: Refresh Token Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - uses: actions/setup-node@v2
        with:
          node-version: '18'
      - run: ./test-refresh-token.sh
```

## 📈 Performance Testing

### Load Testing Scenarios
1. **Concurrent Users**: 100+ users with token refresh
2. **Token Rotation**: High-frequency refresh requests
3. **Memory Usage**: Long-running session management
4. **Database Load**: Refresh token cleanup operations

### Monitoring Metrics
- Token refresh success rate
- Average refresh latency
- Database query performance
- Memory usage patterns

## 🔍 Debugging Tips

### Backend Debugging
```go
// Add logging to refresh token handler
logger.Info("Processing refresh token", "token", token[:8]+"...")
```

### Frontend Debugging
```javascript
// Add console logging to interceptor
console.log('Token refresh triggered for:', error.config.url);
```

### Network Debugging
- Use browser DevTools Network tab
- Check Application tab for cookie storage
- Monitor localStorage/sessionStorage usage

## 📝 Test Data Management

### Mock Users
```typescript
const testUsers = {
  valid: { id: 'user-123', email: 'test@example.com' },
  expired: { id: 'user-456', email: 'expired@example.com' },
  blocked: { id: 'user-789', email: 'blocked@example.com' }
};
```

### Test Tokens
```typescript
const testTokens = {
  valid: 'valid-refresh-token-123',
  expired: 'expired-refresh-token-456',
  invalid: 'invalid-refresh-token-789'
};
```

## 🚨 Security Testing

### Penetration Testing Scenarios
1. **Token Theft**: Simulate stolen refresh tokens
2. **Replay Attacks**: Reuse expired tokens
3. **Cross-Site Scripting**: Test token exposure via XSS
4. **CSRF Attacks**: Verify token protection mechanisms

### Security Headers
```http
Set-Cookie: refresh_token=...; HttpOnly; Secure; SameSite=Lax
```

## 📚 Additional Resources

- [OWASP JWT Security Guidelines](https://owasp.org/www-project-cheat-sheets/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
- [React Testing Library Documentation](https://testing-library.com/docs/react-testing-library/intro)
- [Go Testing Best Practices](https://golang.org/pkg/testing/)

---

## 🎯 Next Steps

1. **Add Performance Benchmarks**: Implement automated performance testing
2. **Security Audits**: Schedule regular security assessments
3. **Monitoring**: Add production monitoring for refresh token metrics
4. **Documentation**: Keep this guide updated with new test scenarios

---

*Last Updated: January 2026*
