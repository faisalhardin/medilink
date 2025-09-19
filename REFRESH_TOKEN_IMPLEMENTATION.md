# Refresh Token Implementation

This document describes the refresh token implementation added to the Medilink authentication system.

## Overview

The refresh token system implements a dual-token approach for enhanced security:
- **Access Token**: Short-lived (15-30 minutes) for API requests
- **Refresh Token**: Long-lived (7-30 days) for obtaining new access tokens

## Features

- ✅ Dual token authentication (access + refresh tokens)
- ✅ Token rotation on refresh
- ✅ **Single device login restriction** - Only one active session per device
- ✅ Device tracking and identification
- ✅ Token revocation (single device and all devices)
- ✅ Automatic cleanup of expired tokens
- ✅ Security headers and IP tracking
- ✅ Database persistence with proper indexing

## API Endpoints

### Authentication Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/v1/auth/refresh` | Refresh access token using refresh token | No |
| POST | `/v1/auth/logout` | Logout single device | No |
| POST | `/v1/auth/logout-all` | Logout all devices | Yes |

### Request/Response Examples

#### Refresh Token Request
```bash
POST /v1/auth/refresh
Content-Type: application/json
X-Device-ID: device-123

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "device_id": "device-123",
  "user_agent": "Mozilla/5.0...",
  "ip_address": "192.168.1.1"
}
```

#### Refresh Token Response
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "new-refresh-token-string",
  "token_type": "Bearer",
  "expires_in": 1800
}
```

#### Logout Request
```bash
POST /v1/auth/logout
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Database Schema

### Refresh Tokens Table
```sql
CREATE TABLE mdl_refresh_tokens (
    id serial4 NOT NULL,
    token varchar(255) UNIQUE NOT NULL,
    user_id int8 NOT NULL,
    institution_id int8 NOT NULL,
    device_id varchar(255),
    user_agent text,
    ip_address varchar(45),
    is_revoked bool DEFAULT false NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    revoked_at timestamptz NULL,
    CONSTRAINT mdl_refresh_tokens_pkey PRIMARY KEY (id)
);
```

### Indexes
- `idx_refresh_tokens_token` - Fast token lookup
- `idx_refresh_tokens_user_id` - User-based queries
- `idx_refresh_tokens_expires_at` - Cleanup operations
- `idx_refresh_tokens_is_revoked` - Revocation queries
- `idx_refresh_tokens_user_device_unique` - **Single device login constraint**

## Configuration

### JWT Configuration
```yaml
jwt_config:
  duration_in_minutes: 30      # Access token duration (15-30 min recommended)
  refresh_duration_in_days: 7  # Refresh token duration (7-30 days recommended)
```

## Security Features

### Token Security
- **Cryptographically Secure**: Refresh tokens use `crypto/rand` for generation
- **Base64 URL Encoding**: Safe for HTTP transmission
- **Unique Constraints**: Database prevents duplicate tokens
- **Expiration**: Automatic token expiration

### Device Tracking
- **Device ID**: Track devices via `X-Device-ID` header
- **User Agent**: Browser/client identification
- **IP Address**: Network location tracking
- **Single Device Login**: Only one active session per device per user

### Token Rotation
- **One-Time Use**: Refresh tokens are revoked after use
- **New Token Pair**: Each refresh generates new access + refresh tokens
- **Prevents Replay**: Old refresh tokens become invalid

### Single Device Login
- **One Session Per Device**: Users can only have one active session per device
- **Automatic Logout**: New login on same device revokes previous session
- **Device Identification**: Uses `X-Device-ID` header to identify devices
- **Database Constraint**: Unique index prevents multiple active tokens per device

## Implementation Details

### File Structure
```
internal/
├── entity/model/auth.go              # Auth models and DTOs
├── repo/auth/
│   ├── refresh_token.go             # Refresh token repository
│   └── jwt.go                       # JWT and token pair creation
├── usecase/auth/auth.go             # Authentication business logic
└── http/auth/handler.go             # HTTP handlers

schema/medianne/
└── add_refresh_tokens.sql           # Database migration
```

### Key Components

#### 1. Token Pair Creation
```go
func (opt *Options) CreateTokenPair(ctx context.Context, payload model.UserJWTPayload, deviceInfo model.DeviceInfo) (tokenPair model.TokenPair, err error)
```

#### 2. Refresh Token Validation
```go
func (opt *Options) GetRefreshToken(ctx context.Context, token string) (*model.RefreshToken, error)
```

#### 3. Token Revocation
```go
func (opt *Options) RevokeRefreshToken(ctx context.Context, token string) error
func (opt *Options) RevokeAllUserRefreshTokens(ctx context.Context, userID int64) error
```

## Single Device Login Behavior

### How It Works
1. **First Login**: User logs in on Device A → Creates new token pair
2. **Second Login on Same Device**: User logs in again on Device A → Previous session is automatically revoked, new token pair created
3. **Login on Different Device**: User logs in on Device B → Both devices can have active sessions (different device IDs)
4. **Token Refresh**: Refresh tokens maintain the same device restriction

### Device Identification
The system uses the `X-Device-ID` header to identify devices. This should be:
- **Unique per device**: Each device should have a unique identifier
- **Persistent**: Same device should use the same ID across sessions
- **Client-generated**: Generated by the client application

### Example Device ID Generation
```javascript
// Generate a persistent device ID
function generateDeviceID() {
  let deviceId = localStorage.getItem('device_id');
  if (!deviceId) {
    deviceId = 'device_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    localStorage.setItem('device_id', deviceId);
  }
  return deviceId;
}
```

## Usage Examples

### Client-Side Implementation

#### 1. Login and Store Tokens
```javascript
// Generate or retrieve device ID
const deviceId = generateDeviceID();

// Login request
const loginResponse = await fetch('/v1/auth/google', {
  method: 'GET',
  headers: {
    'X-Device-ID': deviceId
  }
});

const { access_token, refresh_token } = await loginResponse.json();

// Store tokens securely
localStorage.setItem('access_token', access_token);
localStorage.setItem('refresh_token', refresh_token);
```

#### 2. API Request with Token Refresh
```javascript
async function apiRequest(url, options = {}) {
  let accessToken = localStorage.getItem('access_token');
  
  // Try the request
  let response = await fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': `Bearer ${accessToken}`
    }
  });
  
  // If token expired, refresh and retry
  if (response.status === 401) {
    const newTokens = await refreshAccessToken();
    accessToken = newTokens.access_token;
    
    // Retry the request
    response = await fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        'Authorization': `Bearer ${accessToken}`
      }
    });
  }
  
  return response;
}

async function refreshAccessToken() {
  const refreshToken = localStorage.getItem('refresh_token');
  const deviceId = generateDeviceID();
  
  const response = await fetch('/v1/auth/refresh', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Device-ID': deviceId
    },
    body: JSON.stringify({
      refresh_token: refreshToken
    })
  });
  
  const tokens = await response.json();
  
  // Store new tokens
  localStorage.setItem('access_token', tokens.access_token);
  localStorage.setItem('refresh_token', tokens.refresh_token);
  
  return tokens;
}
```

#### 3. Logout
```javascript
async function logout() {
  const refreshToken = localStorage.getItem('refresh_token');
  
  await fetch('/v1/auth/logout', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      refresh_token: refreshToken
    })
  });
  
  // Clear local storage
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
}
```

## Maintenance

### Cleanup Expired Tokens
```go
// Run this as a scheduled job (e.g., every hour)
func cleanupExpiredTokens() {
    err := authRepo.CleanupExpiredTokens(context.Background())
    if err != nil {
        log.Printf("Failed to cleanup expired tokens: %v", err)
    }
}
```

### Monitoring
- Monitor refresh token usage patterns
- Track failed refresh attempts
- Alert on suspicious activity (multiple devices, unusual IPs)

## Security Considerations

### Best Practices
1. **Store refresh tokens securely** (HttpOnly cookies preferred)
2. **Use HTTPS** for all token-related requests
3. **Implement rate limiting** on refresh endpoint
4. **Monitor for suspicious activity**
5. **Regular token cleanup**

### Token Storage Recommendations
- **Access Tokens**: Memory only (not localStorage)
- **Refresh Tokens**: HttpOnly cookies or secure storage
- **Never store tokens in plain text**

## Migration

### Database Migration
Run the migration script to add the refresh tokens table:
```bash
psql -d your_database -f schema/medianne/add_refresh_tokens.sql
```

### Configuration Update
Update your configuration files to include refresh token settings:
```yaml
jwt_config:
  duration_in_minutes: 30
  refresh_duration_in_days: 7
```

## Testing

### Unit Tests
Test the refresh token functionality:
```go
func TestRefreshToken(t *testing.T) {
    // Test token creation
    // Test token validation
    // Test token revocation
    // Test token rotation
}
```

### Integration Tests
Test the complete flow:
1. Login and receive token pair
2. Use access token for API calls
3. Refresh token when expired
4. Logout and verify token revocation

## Troubleshooting

### Common Issues

1. **Token Not Found**: Check if refresh token exists and is not expired
2. **Invalid Token**: Verify token format and signature
3. **Database Errors**: Check database connection and table existence
4. **Configuration Issues**: Verify JWT settings in config files
5. **Single Device Login Issues**: 
   - Ensure `X-Device-ID` header is sent with all requests
   - Verify device ID is consistent across sessions
   - Check if previous session was properly revoked

### Debug Logging
Enable debug logging to trace token operations:
```go
log.Printf("Refresh token created for user %d", userID)
log.Printf("Token validation result: %v", isValid)
```
