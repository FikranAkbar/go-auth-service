# Missing Features & Improvements

**Last Updated:** January 16, 2026  
**Service:** Go Auth Service v1.0  
**Status:** Core features complete, production enhancements needed

---

## Executive Summary

Your Go Auth Service has a **solid foundation** with core authentication features fully implemented. The architecture is clean, testable, and follows industry best practices. However, several **production-critical features** and **enhancements** are missing before it's ready for real-world deployment.

### Current Completion: ~60%

- ✅ **Core Auth Flow:** Registration, Login, Logout, Email Verification, Token Refresh
- ✅ **Security Basics:** Bcrypt, JWT, Email Verification, Redis Sessions
- ✅ **Architecture:** Clean Architecture, Interface-driven, Testable
- ⚠️ **Missing:** Auth Middleware, Password Reset, Rate Limiting, Monitoring, Complete Tests

---

## 🔴 Critical Missing Features

### 1. Authentication Middleware **[HIGH PRIORITY]**

**Status:** ❌ Not Implemented  
**Impact:** Your protected endpoints are currently unprotected

**Problem:**
- The `/logout` endpoint expects `user_id` from context but there's no middleware to set it
- Comment in `auth_handler.go:138` says "set by auth middleware" but middleware doesn't exist
- Any future protected endpoints (user profile, admin routes) have no auth layer

**What's Needed:**
```go
// File: internal/http/middleware/auth.go
- JWTAuthMiddleware() - validates access token from Authorization header
- ExtractUserFromToken() - parses JWT and sets user_id in context
- RequireAuth() - wrapper middleware for protected routes
```

**Routes That Need Protection:**
- `POST /api/auth/logout` - currently unprotected!
- Future: `GET /api/users/me` - user profile
- Future: `PUT /api/users/me` - update profile
- Future: Any admin endpoints

**Estimated Effort:** 2-3 hours

---

### 2. Password Reset Flow **[HIGH PRIORITY]**

**Status:** ❌ Not Implemented  
**Impact:** Users cannot recover locked-out accounts

**Problem:**
- Users who forget passwords have no recovery mechanism
- This is a **mandatory feature** for any production auth service

**What's Needed:**

**Endpoints:**
```
POST /api/auth/forgot-password
- Input: { "email": "user@example.com" }
- Generates reset token (JWT, 1h expiry)
- Sends email with reset link
- Response: Always 200 (don't reveal if email exists)

POST /api/auth/reset-password
- Input: { "token": "jwt-token", "new_password": "newPass123" }
- Validates reset token
- Updates password
- Invalidates all refresh tokens (logout all sessions)
- Response: 200 with success message
```

**Services Needed:**
```go
// In AuthService:
- RequestPasswordReset(ctx, email) error
- ResetPassword(ctx, token, newPassword) error

// In JWTManager:
- GeneratePasswordResetToken(userID, email) (string, error)

// In EmailService:
- SendPasswordResetEmail(email, resetToken) error
```

**Database Changes:**
- Consider adding `password_reset_requested_at` column to track reset attempts
- Or use Redis to track reset attempts (rate limiting)

**Estimated Effort:** 4-6 hours

---

### 3. Rate Limiting **[HIGH PRIORITY]**

**Status:** ❌ Not Implemented  
**Impact:** Vulnerable to brute force attacks

**Problem:**
- Attackers can attempt unlimited login/registration attempts
- No protection against password guessing
- Email spam vulnerability (verification/reset emails)

**What's Needed:**

**Middleware:**
```go
// File: internal/http/middleware/rate_limit.go
- RateLimitMiddleware(requests, window) - general rate limiter
- LoginRateLimiter() - 5 attempts per 15 minutes per IP
- RegistrationRateLimiter() - 3 registrations per hour per IP
- EmailRateLimiter() - 3 emails per hour per email address
```

**Implementation Options:**
1. **Redis-based** (recommended) - Use Redis INCR with TTL
2. **In-memory** (simple) - Use `golang.org/x/time/rate` limiter
3. **External** (production) - Use Cloudflare, AWS WAF, or Kong

**Rate Limit Rules:**
```
/api/auth/login          - 5 requests / 15 min per IP
/api/auth/register       - 3 requests / 1 hour per IP
/api/auth/resend-verification - 3 requests / 1 hour per email
/api/auth/forgot-password - 3 requests / 1 hour per email
/api/auth/refresh-token  - 10 requests / 1 min per user
```

**Estimated Effort:** 3-4 hours

---

## 🟡 Important Missing Features

### 4. User Profile Management **[MEDIUM PRIORITY]**

**Status:** ⚠️ Handler exists but no endpoints  
**Impact:** Users cannot view/update their profiles

**Problem:**
- `UserHandler` is created but unused (commented out in router)
- No way to get current user info
- No way to update email/username
- No way to change password

**What's Needed:**

**Endpoints:**
```
GET /api/users/me
- Returns current user profile
- Requires auth middleware

PUT /api/users/me
- Update username (with validation)
- Requires auth middleware
- Re-validate uniqueness

PUT /api/users/me/email
- Request email change (send verification to new email)
- Old email remains active until new email verified
- Requires auth middleware

PUT /api/users/me/password
- Change password (requires current password)
- Invalidates all refresh tokens
- Requires auth middleware

DELETE /api/users/me
- Soft delete user account
- Or hard delete with confirmation
- Requires auth middleware
```

**Services Needed:**
```go
// In UserService:
- GetCurrentUser(ctx, userID) (*User, error)
- UpdateUsername(ctx, userID, newUsername) error
- RequestEmailChange(ctx, userID, newEmail) error
- ChangePassword(ctx, userID, currentPassword, newPassword) error
- DeleteAccount(ctx, userID) error
```

**Estimated Effort:** 5-6 hours

---

### 5. Complete Test Coverage **[MEDIUM PRIORITY]**

**Status:** ⚠️ Partial (only services tested)  
**Impact:** Cannot confidently deploy or refactor

**Current Coverage:**
```
✅ internal/service/auth_service_test.go - EXISTS
✅ internal/service/user_service_test.go - EXISTS
✅ internal/config/config_test.go - EXISTS
❌ internal/handler/* - NO TESTS
❌ internal/repository/* - NO TESTS
❌ internal/security/* - NO TESTS
❌ pkg/* - NO TESTS
```

**What's Needed:**

**Handler Tests (Integration):**
```go
// internal/handler/auth_handler_test.go
- TestRegister_Success
- TestRegister_InvalidEmail
- TestRegister_DuplicateEmail
- TestLogin_Success
- TestLogin_InvalidCredentials
- TestLogin_UnverifiedEmail
- TestVerifyEmail_Success
- TestVerifyEmail_ExpiredToken
- TestRefreshToken_Success
- TestRefreshToken_InvalidToken
- TestLogout_Success
```

**Repository Tests (Database):**
```go
// internal/repository/user_repository_test.go
- TestCreateUser
- TestGetUserByID
- TestGetUserByEmail
- TestExistsByEmail
- TestUpdateUser
- Test with real database (testcontainers)

// internal/repository/token_repository_test.go
- TestStoreRefreshToken
- TestValidateRefreshToken
- TestDeleteRefreshToken
- Test with real Redis (testcontainers)
```

**Security Tests:**
```go
// internal/security/jwt_manager_test.go
- TestGenerateAccessToken
- TestGenerateRefreshToken
- TestValidateToken_Expired
- TestValidateToken_InvalidSignature

// internal/security/password_hasher_test.go
- TestHashPassword
- TestVerifyPassword
```

**Target Coverage:** 80%+

**Estimated Effort:** 8-10 hours

---

### 6. API Documentation (Swagger/OpenAPI) **[MEDIUM PRIORITY]**

**Status:** ⚠️ Manual docs only  
**Impact:** Hard for frontend/mobile devs to integrate

**Problem:**
- Only manual markdown docs in `docs/API_ENDPOINTS.md`
- No interactive API playground
- No auto-generated client SDKs
- Hard to keep docs in sync with code

**What's Needed:**

**Implementation:**
```go
// Use: github.com/swaggo/swag

// Add swagger comments to handlers:
// @Summary Register new user
// @Description Creates a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} RegisterResponse
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

**Endpoints to Add:**
```
GET /swagger/index.html - Swagger UI
GET /swagger/doc.json - OpenAPI spec
```

**Benefits:**
- Interactive API testing
- Auto-generated documentation
- Client SDK generation
- API versioning support

**Estimated Effort:** 4-5 hours (initial setup + documenting all endpoints)

---

## 🟢 Nice-to-Have Features

### 7. Monitoring & Metrics **[LOW PRIORITY]**

**Status:** ❌ Not Implemented  
**Impact:** No visibility into service health/performance

**What's Needed:**

**Prometheus Metrics:**
```go
// Metrics to track:
- auth_register_total (counter)
- auth_login_total (counter)
- auth_login_failures (counter)
- auth_token_refresh_total (counter)
- http_request_duration_seconds (histogram)
- redis_connection_errors (counter)
- database_connection_errors (counter)
- email_send_total (counter)
- email_send_failures (counter)
```

**Health Check Enhancement:**
```
GET /health/live - Kubernetes liveness probe
GET /health/ready - Kubernetes readiness probe (checks DB, Redis)
GET /metrics - Prometheus metrics endpoint
```

**Estimated Effort:** 3-4 hours

---

### 8. Session Management **[LOW PRIORITY]**

**Status:** ⚠️ Basic (only logout current session)  
**Impact:** Users can't manage active sessions

**What's Needed:**

**Endpoints:**
```
GET /api/users/me/sessions
- List all active sessions
- Shows: device, IP, location, last active

DELETE /api/users/me/sessions/{sessionId}
- Revoke specific session

DELETE /api/users/me/sessions
- Logout all devices (except current)
```

**Redis Structure:**
```
user:{userID}:sessions:{sessionID} = {
  refresh_token: "...",
  ip_address: "192.168.1.1",
  user_agent: "Mozilla/5.0...",
  created_at: "2026-01-16T10:00:00Z",
  last_used_at: "2026-01-16T12:30:00Z"
}
```

**Estimated Effort:** 4-5 hours

---

### 9. Email Templates (HTML) **[LOW PRIORITY]**

**Status:** ⚠️ Plain text only  
**Impact:** Emails look unprofessional

**Current:**
```
Subject: Verify Your Email
Body: Click here: http://localhost:8080/verify?token=xxx
```

**What's Needed:**
```go
// Use: html/template

// Templates:
- templates/verification_email.html
- templates/password_reset_email.html
- templates/password_changed_email.html (notification)
- templates/login_notification_email.html (security alert)
```

**Features:**
- Branded HTML design
- Responsive (mobile-friendly)
- Call-to-action buttons
- Company logo
- Unsubscribe link (if applicable)

**Estimated Effort:** 3-4 hours

---

### 10. Request Validation (Structured) **[LOW PRIORITY]**

**Status:** ⚠️ Manual validation in services  
**Impact:** Inconsistent error messages

**Current:**
```go
// Validation scattered across services
if email == "" {
    return errors.New("email required")
}
```

**What's Needed:**
```go
// Use: github.com/go-playground/validator

type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email,max=100"`
    Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
    Password string `json:"password" validate:"required,min=8,max=128"`
}

// In handler:
if err := validator.Validate(req); err != nil {
    return ValidationErrors(err)
}
```

**Benefits:**
- Consistent validation rules
- Better error messages
- Less boilerplate code
- Centralized validation logic

**Estimated Effort:** 2-3 hours

---

### 11. Logging Enhancements **[LOW PRIORITY]**

**Status:** ⚠️ Basic logging exists  
**Impact:** Hard to debug production issues

**Current:**
```go
logger.Info("User registered")
logger.Errorf("Failed: %v", err)
```

**What's Needed:**
```go
// Structured logging with context:
logger.WithFields(map[string]interface{}{
    "user_id": userID,
    "email": email,
    "ip_address": r.RemoteAddr,
    "request_id": requestID,
}).Info("User registered successfully")

// Trace IDs for distributed tracing
// Log levels per environment (debug in dev, info in prod)
// Log sampling (don't log every success in prod)
```

**Integration:**
- OpenTelemetry tracing
- Log aggregation (ELK, Datadog, CloudWatch)
- Error tracking (Sentry)

**Estimated Effort:** 3-4 hours

---

### 12. Environment-based Configuration **[LOW PRIORITY]**

**Status:** ⚠️ Only `config.yaml`  
**Impact:** Manual config changes per environment

**Current:**
```yaml
# config.yaml
jwt:
  secret_key: gZ/eZd24Iv8JfKG075iSTx3BvHFd5ko5qKOAK0Lxlao=
```

**What's Needed:**
```
config/
  config.yaml          # Base config
  config.dev.yaml      # Development overrides
  config.staging.yaml  # Staging overrides
  config.prod.yaml     # Production overrides

# Support environment variables:
JWT_SECRET_KEY=xxx
DB_PASSWORD=xxx
REDIS_PASSWORD=xxx
```

**Benefits:**
- Secrets management (AWS Secrets Manager, Vault)
- Different configs per environment
- 12-factor app compliance

**Estimated Effort:** 2-3 hours

---

## 📊 Prioritization Matrix

### Phase 1: Production-Ready (Must Have)
**Estimated: 15-20 hours**

1. ✅ Authentication Middleware (2-3h) - **CRITICAL**
2. ✅ Password Reset Flow (4-6h) - **CRITICAL**
3. ✅ Rate Limiting (3-4h) - **CRITICAL**
4. ✅ Complete Tests for Handlers (6-8h) - **HIGH**

**Goal:** Service is secure and minimally viable for production

---

### Phase 2: User Experience (Should Have)
**Estimated: 15-18 hours**

5. ✅ User Profile Management (5-6h)
6. ✅ API Documentation (4-5h)
7. ✅ Complete Tests for Repos (4-5h)
8. ✅ HTML Email Templates (3-4h)

**Goal:** Service is user-friendly and well-documented

---

### Phase 3: Operational Excellence (Nice to Have)
**Estimated: 12-15 hours**

9. ✅ Monitoring & Metrics (3-4h)
10. ✅ Session Management (4-5h)
11. ✅ Request Validation (2-3h)
12. ✅ Logging Enhancements (3-4h)

**Goal:** Service is observable and maintainable

---

## 🎯 Recommended Next Steps

### Option A: Quick Production Deploy (Phase 1 Only)
**Timeline:** 1-2 weeks  
**Focus:** Security + stability  
**Deploy to:** Fly.io / Railway / Render  
**Best for:** MVP, proof of concept, learning

### Option B: Polished Product (Phase 1 + 2)
**Timeline:** 3-4 weeks  
**Focus:** Security + UX  
**Deploy to:** Any cloud provider  
**Best for:** Real users, portfolio project

### Option C: Enterprise-Grade (All Phases)
**Timeline:** 5-6 weeks  
**Focus:** Everything  
**Deploy to:** AWS / GCP / Azure with monitoring  
**Best for:** Production SaaS, commercial product

---

## 🔧 Technical Debt

### Known Issues

1. **Logout endpoint currently unprotected** - needs auth middleware
2. **UserHandler created but unused** - no user management endpoints
3. **No repository tests** - database layer untested
4. **Email send failures after user registration** - user created but can't login (current design commits first)
5. **Bcrypt cost = 10** - 3 seconds per request (may need tuning for prod)
6. **No CORS configuration** - will fail with frontend apps
7. **No graceful shutdown for Redis** - only DB is properly closed
8. **Hard-coded token expiries** - should match config values

---

## 📈 Estimated Total Effort

| Phase | Hours | Developer Days (6h/day) |
|-------|-------|------------------------|
| Phase 1 (Must Have) | 15-20h | 3-4 days |
| Phase 2 (Should Have) | 15-18h | 3 days |
| Phase 3 (Nice to Have) | 12-15h | 2-3 days |
| **TOTAL** | **42-53h** | **8-10 days** |

*Note: Estimates assume one developer working part-time (6 hours/day)*

---

## 🚀 When to Deploy?

### ✅ Safe to Deploy After Phase 1 IF:
- You add auth middleware
- You add rate limiting
- You add password reset
- You deploy over HTTPS (TLS required!)
- You change JWT secret from example
- You use strong DB/Redis passwords
- You enable CORS for your frontend

### ⚠️ Don't Deploy Yet IF:
- Endpoints are still unprotected
- No rate limiting (easy DDoS target)
- Using example secrets in config

---

## 📝 Notes

- This service is **well-architected** - clean, testable, maintainable
- The **foundation is solid** - adding features will be straightforward
- You've avoided **common pitfalls** - transactions, interfaces, error handling
- The **biggest gap** is production hardening (auth middleware, rate limits, monitoring)

**You're 60% done with a very clean 60%.** The remaining 40% is mostly "wrapping paper" to make it production-ready and user-friendly.

---

**Last Updated:** January 16, 2026  
**Next Review:** After implementing Phase 1 features

