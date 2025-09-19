# Single Device Login Test Guide

This guide demonstrates how to test the single device login functionality.

## Test Scenarios

### Scenario 1: First Login on Device
1. **Send login request with device ID:**
   ```bash
   curl -X GET "http://localhost:8080/v1/auth/google" \
     -H "X-Device-ID: device-123"
   ```

2. **Expected Result:**
   - Returns new token pair (access_token + refresh_token)
   - No existing tokens to revoke

### Scenario 2: Second Login on Same Device
1. **Send another login request with same device ID:**
   ```bash
   curl -X GET "http://localhost:8080/v1/auth/google" \
     -H "X-Device-ID: device-123"
   ```

2. **Expected Result:**
   - Previous session is automatically revoked
   - New token pair is created
   - Old refresh token becomes invalid

### Scenario 3: Login on Different Device
1. **Send login request with different device ID:**
   ```bash
   curl -X GET "http://localhost:8080/v1/auth/google" \
     -H "X-Device-ID: device-456"
   ```

2. **Expected Result:**
   - Both devices can have active sessions
   - No tokens are revoked (different device IDs)

### Scenario 4: Token Refresh with Device Change
1. **Refresh token with different device ID:**
   ```bash
   curl -X POST "http://localhost:8080/v1/auth/refresh" \
     -H "Content-Type: application/json" \
     -H "X-Device-ID: device-789" \
     -d '{
       "refresh_token": "your-refresh-token"
     }'
   ```

2. **Expected Result:**
   - Old tokens for device-789 are revoked
   - New token pair is created for device-789
   - Device-123 session remains active

## Database Verification

### Check Active Tokens
```sql
-- Check all active tokens for a user
SELECT user_id, device_id, token, created_at, expires_at 
FROM mdl_refresh_tokens 
WHERE user_id = 1 AND is_revoked = false AND expires_at > NOW()
ORDER BY created_at DESC;
```

### Check Revoked Tokens
```sql
-- Check revoked tokens
SELECT user_id, device_id, token, revoked_at 
FROM mdl_refresh_tokens 
WHERE user_id = 1 AND is_revoked = true
ORDER BY revoked_at DESC;
```

## Client-Side Testing

### JavaScript Test Function
```javascript
// Test single device login
async function testSingleDeviceLogin() {
  const deviceId = 'test-device-' + Date.now();
  
  console.log('Testing single device login...');
  
  // First login
  console.log('1. First login on device:', deviceId);
  const login1 = await login(deviceId);
  console.log('Login 1 tokens:', login1);
  
  // Second login on same device
  console.log('2. Second login on same device:', deviceId);
  const login2 = await login(deviceId);
  console.log('Login 2 tokens:', login2);
  
  // Try to use old refresh token (should fail)
  console.log('3. Testing old refresh token...');
  try {
    await refreshToken(login1.refresh_token, deviceId);
    console.log('❌ Old token should have been revoked!');
  } catch (error) {
    console.log('✅ Old token properly revoked:', error.message);
  }
  
  // Use new refresh token (should work)
  console.log('4. Testing new refresh token...');
  try {
    const newTokens = await refreshToken(login2.refresh_token, deviceId);
    console.log('✅ New token works:', newTokens);
  } catch (error) {
    console.log('❌ New token failed:', error.message);
  }
}

async function login(deviceId) {
  const response = await fetch('/v1/auth/google', {
    method: 'GET',
    headers: {
      'X-Device-ID': deviceId
    }
  });
  return await response.json();
}

async function refreshToken(refreshToken, deviceId) {
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
  return await response.json();
}

// Run the test
testSingleDeviceLogin();
```

## Expected Behavior Summary

| Action | Device ID | Expected Result |
|--------|-----------|-----------------|
| Login | device-123 | ✅ New session created |
| Login | device-123 | ✅ Previous session revoked, new session created |
| Login | device-456 | ✅ New session created (different device) |
| Refresh | device-123 | ✅ Tokens refreshed for device-123 |
| Refresh | device-456 | ✅ Tokens refreshed for device-456 |
| Refresh | device-789 | ✅ Old device-789 tokens revoked, new tokens created |

## Troubleshooting

### Common Issues

1. **Multiple sessions on same device:**
   - Check if `X-Device-ID` header is being sent consistently
   - Verify device ID format is the same across requests

2. **Token not revoked:**
   - Check database for revoked tokens
   - Verify the revoke query is working correctly

3. **Database constraint violations:**
   - Check if the unique index is properly created
   - Verify the WHERE clause in the unique index

### Debug Queries

```sql
-- Check unique constraint
SELECT constraint_name, constraint_type 
FROM information_schema.table_constraints 
WHERE table_name = 'mdl_refresh_tokens';

-- Check indexes
SELECT indexname, indexdef 
FROM pg_indexes 
WHERE tablename = 'mdl_refresh_tokens';

-- Test unique constraint
INSERT INTO mdl_refresh_tokens (user_id, device_id, token, expires_at) 
VALUES (1, 'test-device', 'test-token', NOW() + INTERVAL '1 day');
-- This should fail if constraint is working
```
