# Security considerations

The goal of this project is to be able to host documents and make them accessible via a token (or other auth method)
while disabling public access, listing, and indexing by search engines.

## Implemented Security Measures

### ✅ **Token Security**
- **No token logging**: Token validation errors log only the requested path, not token details
- **Multiple token sources**: Query parameter, Authorization header, or cookie
- **Secure validation**: HMAC-SHA256 signatures with timing-safe comparison
- **Configurable expiration**: Tokens expire automatically

### ✅ **Search Engine Prevention**
- **X-Robots-Tag**: `noindex, nofollow, noarchive, nosnippet` header on all document responses
- **Referrer-Policy**: `no-referrer` header prevents referrer leakage
- **Private documents**: All documents require authentication, preventing indexing

### ✅ **Directory Listing Prevention**
- **Explicit blocking**: Requests ending with `/` return `403 Forbidden`
- **No directory browsing**: Only specific files can be accessed
- **Default file handling**: Empty paths default to `index.html` only

### ✅ **Security Headers**
- **X-Content-Type-Options**: `nosniff` prevents MIME type sniffing attacks
- **X-Frame-Options**: `SAMEORIGIN` provides clickjacking protection
- **Referrer-Policy**: `no-referrer` prevents referrer information leakage
- **Content-Type**: Proper MIME type detection and setting

### ✅ **Caching Policy**
- **HTML files**: `private, max-age=60` (short-lived, private cache)
- **Assets** (CSS/JS/images): `private, max-age=3600` (longer cache, still private)
- **Authentication required**: All caching respects token-based access

### ✅ **Iframe Security**
- **referrerpolicy="no-referrer"**: Automatically added to generated iframes
- **Configurable sandbox**: Support for iframe sandbox restrictions
- **Token integration**: Secure token embedding in iframe URLs

### ✅ **Container Security**
- **Non-root execution**: Container runs as unprivileged user (appuser:1001)
- **Minimal attack surface**: Alpine Linux base with only necessary packages
- **Health checks**: Built-in container health monitoring

## Additional Recommendations

### 🔄 **Optional Enhancements**
- **Content-Security-Policy**: Add CSP header for specific iframe embedding domains
  ```
  Content-Security-Policy: frame-ancestors https://your-lms.example.com
  ```
- **Short token parameter**: Consider using `t` instead of `token` to reduce URL length
- **Token rotation**: Implement periodic token secret rotation
- **Rate limiting**: Add request rate limiting for production deployments

### 🛡️ **Deployment Security**
- **HTTPS only**: Cloud Run provides HTTPS by default
- **Private bucket**: Google Cloud Storage bucket is private, not publicly accessible
- **IAM restrictions**: Service account has minimal required permissions (Storage Object Viewer)
- **Environment variables**: Sensitive configuration via environment variables, not files

## Security Testing

The implementation includes security tests that verify:
- Directory listing prevention returns 403 Forbidden
- Security headers are properly set
- Token validation doesn't leak sensitive information

## Threat Model Coverage

| Threat | Mitigation |
|--------|------------|
| **Unauthorized access** | Token-based authentication required |
| **Token theft** | Short expiration times, secure signatures |
| **Search engine indexing** | X-Robots-Tag headers, private authentication |
| **Directory listing** | Explicit 403 blocking for directory paths |
| **Information leakage** | Generic error messages, no token logging |
| **Clickjacking** | X-Frame-Options header |
| **MIME sniffing** | X-Content-Type-Options header |
| **Referrer leakage** | Referrer-Policy headers and iframe attributes |
| **Container compromise** | Non-root execution, minimal attack surface |

This security implementation ensures documents remain private, non-indexable, and accessible only through authenticated tokens while preventing common web security vulnerabilities.
