# Google Sign-In — Android (Kotlin) + Go Backend

Production-ready implementation using:
- **Android**: Credential Manager API (Google's current recommendation)
- **Backend**: Go with JWT access/refresh token rotation + PostgreSQL

---

## Architecture

```
Android App
  │
  │  1. User taps "Sign in with Google"
  │  2. Credential Manager shows bottom sheet / account picker
  │  3. Google returns a signed ID token
  │
  ▼
POST /api/v1/auth/google   { id_token, nonce }
  │
  │  4. Go backend verifies ID token with Google's public keys
  │  5. Upserts user in PostgreSQL
  │  6. Issues access JWT (15 min) + refresh JWT (7 days)
  │
  ▼
Android stores tokens in EncryptedSharedPreferences
  │
  │  7. Access token is attached to every API request via OkHttp interceptor
  │  8. On 401, TokenAuthenticator calls POST /api/v1/auth/refresh automatically
  │  9. Refresh tokens are rotated on every use (old one revoked)
```

---

## Android Setup

### 1. Google Cloud Console

1. Go to https://console.cloud.google.com
2. Create or select a project
3. Navigate to **APIs & Services → Credentials → + Create Credentials → OAuth 2.0 Client ID**
4. Create an **Android** client:
  - Package name: `com.yourapp` (match your app)
  - SHA-1: run `./gradlew signingReport` to get it
5. Create a **Web application** client — copy this Client ID as `WEB_CLIENT_ID`
6. Download `google-services.json` and place it in `app/`

### 2. Update constants

In `AuthViewModel.kt`:
```kotlin
const val WEB_CLIENT_ID = "YOUR_WEB_CLIENT_ID.apps.googleusercontent.com"
```

In `NetworkModule.kt`:
```kotlin
private const val BASE_URL = "https://api.yourapp.com/"
```

### 3. Add Google logo asset

Place a Google "G" logo SVG/PNG at `app/src/main/res/drawable/ic_google_logo.xml`
(download from https://developers.google.com/identity/branding-guidelines)

---

## Go Backend Setup

### 1. Database

```bash
psql $DATABASE_URL -f migrations/001_init.sql
```

### 2. Environment

```bash
cp .env.example .env
# Fill in all values — especially GOOGLE_CLIENT_ID and JWT_SECRET
```

### 3. Run

```bash
go mod download
go run ./main.go
```

### 4. Test

```bash
go test ./...
```

---

## API Reference

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/google` | None | Exchange Google ID token for JWT pair |
| POST | `/api/v1/auth/refresh` | None | Rotate refresh token, get new access token |
| POST | `/api/v1/auth/signout` | Bearer | Revoke all refresh tokens |
| GET  | `/api/v1/me` | Bearer | Get current user profile |

---

## Security Notes

- **Never** trust the ID token from the client for authorization — always verify server-side
- Nonce prevents replay attacks (Android generates, backend validates)
- Access tokens are short-lived (15 min); refresh tokens are rotated on every use
- Refresh tokens are stored in DB, enabling server-side revocation (sign-out from all devices)
- `EncryptedSharedPreferences` uses Android Keystore — tokens are hardware-backed on supported devices
- JWT signing uses HMAC-SHA256; rotate `JWT_SECRET` periodically in production

---

## Dependencies

### Android
- `androidx.credentials:credentials` — Credential Manager API
- `com.google.android.libraries.identity.googleid` — Google ID option builders
- `androidx.security:security-crypto` — EncryptedSharedPreferences
- `retrofit2` + `moshi` — networking
- `hilt` — dependency injection

### Go
- `github.com/go-chi/chi/v5` — HTTP router
- `github.com/golang-jwt/jwt/v5` — JWT
- `google.golang.org/api/idtoken` — Google ID token verification
- `github.com/lib/pq` — PostgreSQL driver
