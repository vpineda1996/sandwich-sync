# Scotia Authentication Flow

## Required Tokens
- `session-id={SESSION_ID}`
- Akamai cookies

## Authentication Flow

### 1. Initial Request
```
GET https://secure.scotiabank.com/accounts?lng=en&intcmp=S1IORI0621-003&fromLink=true
```
- Results in 302 redirect

### 2. Authorization Request
```
GET https://secure.scotiabank.com/auth/authorize?state={SOME_JWT_TOKEN}&language=en
```
- Forwards to auth with OAuth key

### 3. Authentication
```
GET https://auth.scotiaonline.scotiabank.com/online?oauth_key={SMALL_KEY_NAME}&oauth_key_signature={SOME_JWT_TOKEN}
```

**Two possible outcomes:**
1. If correct cookie present:
  - JWT gets signed
  - 302 redirect to secure.scotia
2. If cookie missing:
  - Login page appears

### 4. User Authentication (if cookie missing)
```
POST https://auth.scotiaonline.scotiabank.com/v2/authentications
```
- Send key and signature in headers with extra information
- Returns KEY for challenge response URL

### 5. Challenge Response
```
POST https://auth.scotiaonline.scotiabank.com/v2/authentications/{KEY}
```
- Send password (once only)
- Poll until challenge resolved
- After successful login, receive redirect URL with state and value
- Important: Capture `x-auth-token` header (this is the LOGIN_ID)

### 6. Get Persistent 2FA Cookie
```
GET https://auth.scotiaonline.scotiabank.com/api/multi-user/{USER_WITH_STARS_IN_MIDDLE}
```
- User ID from `bns-auth-saved-users` cookie
- Returns user info and persistent cookie for future requests

### 7. Final Authorization
```
GET https://secure.scotiabank.com/auth/authorization?code={SOME_CODE}&state={SOME_JWT_STATE}&lng=en&log_id={LOGIN_ID}
```
- Receive essential cookies:
  - `session-id`
  - `bm_sz`
  - `bm_sv`

### 8. Completion
```
GET https://secure.scotiabank.com/accounts
```
- Final redirect to original URL