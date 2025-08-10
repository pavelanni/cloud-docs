# Cloud Docs - Product Requirements Document

**Version**: 1.0  
**Date**: August 2025  
**Status**: Implemented  

## Executive Summary

Cloud Docs is a serverless document hosting platform that provides secure, token-based access to HTML documents and their associated static assets. The platform enables organizations to host private documentation that can be embedded in learning management systems (LMS) or other applications while maintaining strict access control.

## Problem Statement

### Current Challenges
- **Public hosting security**: Traditional web hosting exposes documents to unauthorized access
- **Search engine indexing**: Private documents get indexed by search engines
- **Complex authentication**: Existing solutions require complex user management systems
- **LMS integration difficulty**: Hard to embed private content in learning platforms
- **Static asset complexity**: CSS/JS requires token management, breaking standard HTML development

### Target Users
- **Corporate training departments** hosting private educational content
- **Software companies** providing customer documentation
- **Educational institutions** with premium/private course materials
- **Compliance organizations** requiring access-controlled document distribution

## Solution Overview

Cloud Docs provides a two-tier security model:
1. **Protected documents** (HTML, sensitive content) require authentication tokens
2. **Public static assets** (CSS, JavaScript, images) are served without tokens for standard web development

### Key Benefits
- ✅ **Secure by default**: All sensitive content requires valid tokens
- ✅ **No search engine indexing**: X-Robots-Tag headers prevent crawling
- ✅ **Standard HTML development**: CSS/JS assets work with normal `<link>` and `<script>` tags
- ✅ **Serverless scalability**: Auto-scales from 0 to thousands of concurrent users
- ✅ **Easy LMS integration**: Perfect for iframe embedding
- ✅ **Zero user management**: Stateless token-based authentication

## Functional Requirements

### Core Features

#### 1. Document Hosting
- **FR-1.1**: Host HTML documents on Google Cloud Storage
- **FR-1.2**: Preserve directory structure during upload
- **FR-1.3**: Support all web assets (HTML, CSS, JS, images, PDFs)
- **FR-1.4**: Automatic MIME type detection and appropriate Content-Type headers

#### 2. Two-Tier Security Model
- **FR-2.1**: Protected document routes (`/docs/*`) require valid authentication tokens
- **FR-2.2**: Public static asset routes (`/docs/static/*`) accessible without tokens
- **FR-2.3**: Directory listing prevention (403 Forbidden for paths ending in `/`)
- **FR-2.4**: Automatic `index.html` fallback for root document requests

#### 3. Authentication System
- **FR-3.1**: Custom JWT-inspired token format with HMAC-SHA256 signatures
- **FR-3.2**: Configurable token expiration (minutes to days)
- **FR-3.3**: Multiple token input methods (query parameter, Authorization header, cookie)
- **FR-3.4**: Stateless validation (no database or session storage required)
- **FR-3.5**: Unique request tracking with UUID-based token IDs

#### 4. Security Headers
- **FR-4.1**: `X-Robots-Tag: noindex, nofollow, noarchive, nosnippet` on all documents
- **FR-4.2**: `X-Content-Type-Options: nosniff` to prevent MIME sniffing attacks
- **FR-4.3**: `X-Frame-Options: SAMEORIGIN` for clickjacking protection
- **FR-4.4**: `Referrer-Policy: no-referrer` to prevent information leakage

#### 5. Caching Strategy
- **FR-5.1**: Private caching for documents (`Cache-Control: private, max-age=60`)
- **FR-5.2**: Public caching for static assets (`Cache-Control: public, max-age=3600`)
- **FR-5.3**: Appropriate cache headers based on content type

#### 6. Management Tools
- **FR-6.1**: Upload CLI tool for bulk document deployment
- **FR-6.2**: Token generation CLI tool with configurable expiration
- **FR-6.3**: Token validation CLI tool for debugging
- **FR-6.4**: Iframe generation CLI tool with automatic token embedding

### Advanced Features

#### 7. Upload Tool Capabilities
- **FR-7.1**: Recursive directory upload with structure preservation
- **FR-7.2**: Smart exclusion patterns (`.git/*`, `*.tmp`, `*.log`, `.DS_Store`)
- **FR-7.3**: Custom exclusion pattern support
- **FR-7.4**: Progress reporting during upload
- **FR-7.5**: Dry-run mode for testing uploads
- **FR-7.6**: Upload statistics (files uploaded, skipped, total bytes, duration)

#### 8. Iframe Generation
- **FR-8.1**: Automatic token generation or use provided token
- **FR-8.2**: Configurable iframe attributes (width, height, sandbox, etc.)
- **FR-8.3**: Security attributes (`referrerpolicy="no-referrer"`)
- **FR-8.4**: URL encoding for token safety
- **FR-8.5**: Custom HTML attributes support

#### 9. Operational Features
- **FR-9.1**: Health check endpoints for monitoring (`/health`, `/ping`)
- **FR-9.2**: Structured JSON health responses with timestamps
- **FR-9.3**: Graceful shutdown handling
- **FR-9.4**: Request logging with unique request IDs
- **FR-9.5**: Error logging without token exposure

## Technical Requirements

### Architecture
- **TR-1**: Serverless deployment on Google Cloud Run
- **TR-2**: Go 1.24+ with Chi v5 HTTP router
- **TR-3**: Google Cloud Storage for document storage
- **TR-4**: Multi-stage Docker builds for optimal container size
- **TR-5**: Alpine Linux base image with security hardening

### Performance
- **TR-6**: Cold start time < 3 seconds
- **TR-7**: Warm response time < 100ms for documents
- **TR-8**: Static asset response time < 50ms
- **TR-9**: Support for concurrent requests (80 per instance)
- **TR-10**: Auto-scaling from 0 to 1000+ instances

### Security
- **TR-11**: Non-root container execution (user 1001)
- **TR-12**: HTTPS/TLS 1.3 by default (Cloud Run provided)
- **TR-13**: Timing-safe token signature comparison
- **TR-14**: No sensitive information in error messages or logs
- **TR-15**: Minimal container attack surface

### Scalability
- **TR-16**: Stateless application design (no database required)
- **TR-17**: Horizontal scaling via Cloud Run
- **TR-18**: Support for multiple regions
- **TR-19**: CDN-ready with proper cache headers

## Non-Functional Requirements

### Reliability
- **NFR-1**: 99.9% uptime (Cloud Run SLA)
- **NFR-2**: Automatic instance recovery
- **NFR-3**: Graceful degradation during errors
- **NFR-4**: Circuit breaker pattern for external dependencies

### Security
- **NFR-5**: No data stored in application (stateless)
- **NFR-6**: Token secrets managed via environment variables
- **NFR-7**: Audit trail via Cloud Logging
- **NFR-8**: Regular security updates via base image updates

### Usability
- **NFR-9**: Standard HTML development workflow
- **NFR-10**: Single-command deployment
- **NFR-11**: Clear documentation and examples
- **NFR-12**: Intuitive CLI tool interfaces

### Maintainability
- **NFR-13**: Comprehensive test coverage (>90%)
- **NFR-14**: Clear separation of concerns
- **NFR-15**: Docker-based development environment
- **NFR-16**: Automated build and deployment pipeline

## API Specification

### Public Endpoints
```
GET /health          - Health check (JSON response)
GET /ping           - Simple health check (text response)
GET /               - Service identification
```

### Static Asset Endpoints (No Authentication)
```
GET /docs/static/{path}  - Serve CSS, JS, images without tokens
```

### Protected Endpoints (Token Required)
```
GET /docs/{path}?token={token}  - Serve documents with authentication
GET /docs/              - Serve index.html with authentication
```

### Token Format
```
{base64-payload}.{base64-signature}

Payload: {"id":"uuid","expires_at":"iso-date","issued_at":"iso-date"}
Signature: HMAC-SHA256 of payload using secret key
```

## CLI Tools Specification

### Upload Tool
```bash
./bin/upload -source ./docs -bucket my-bucket [-prefix v1] [-exclude "*.tmp,*.log"]
```

### Token Tool
```bash
./bin/token -generate [-expires 24h]
./bin/token -validate "token-string"
```

### Iframe Tool
```bash
./bin/iframe -document "/path/doc.html" -base-url "https://service-url" 
             [-width "100%"] [-height "600"] [-sandbox "allow-scripts"]
```

## Deployment Requirements

### Environment Variables
- `PORT`: HTTP server port (default: 8080)
- `BUCKET_NAME`: Google Cloud Storage bucket name (required)
- `TOKEN_SECRET`: HMAC signing secret (required)
- `DOCS_PATH`: URL path prefix (default: /docs)
- `LOG_LEVEL`: Logging verbosity (default: info)

### GCP Resources Required
- Google Cloud Project with billing enabled
- Cloud Storage bucket for document storage
- Cloud Run service for application hosting
- Container Registry for Docker images
- IAM service account with Storage Object Viewer role

### Document Organization
```
documents/
├── index.html              # Protected (requires token)
├── docs/                   # Protected (requires token)
│   ├── guide.html
│   └── reference.html
└── static/                 # Public (no token required)
    ├── main.css
    ├── app.js
    └── images/
        └── logo.png
```

## Success Metrics

### Technical Metrics
- Response time < 100ms (95th percentile)
- Error rate < 0.1%
- Token validation time < 5ms
- Container start time < 3 seconds

### Business Metrics
- Zero unauthorized document access
- Zero search engine indexing of protected content
- 100% HTML compatibility (standard CSS/JS linking works)
- Successful iframe embedding in target LMS platforms

## Testing Strategy

### Unit Tests
- Token generation and validation
- Authentication middleware
- File serving logic
- CLI tool functionality

### Integration Tests
- End-to-end document serving
- Static asset serving
- Token authentication flows
- Error handling scenarios

### Security Tests
- Token tampering attempts
- Directory traversal prevention
- Header injection prevention
- Timing attack resistance

### Performance Tests
- Load testing with concurrent requests
- Cold start performance
- Large file serving performance
- Token validation under load

## Risks and Mitigations

### Technical Risks
- **Risk**: Cloud Run domain changes
- **Mitigation**: Document domain and make configurable in consuming systems

- **Risk**: GCS bucket access issues
- **Mitigation**: Proper IAM configuration and monitoring

- **Risk**: Token secret compromise
- **Mitigation**: Regular secret rotation and secure environment variable management

### Security Risks
- **Risk**: Static asset exposure
- **Mitigation**: Only non-sensitive assets in /static/, clear documentation guidelines

- **Risk**: Token leakage in logs
- **Mitigation**: Careful logging implementation, token-free error messages

### Operational Risks
- **Risk**: Vendor lock-in to GCP
- **Mitigation**: Standard Docker deployment, documented architecture

## Future Enhancements

### Phase 2 Potential Features
- Content-Security-Policy headers for specific iframe domains
- Rate limiting for production deployments
- Token usage analytics and reporting
- Multi-region deployment support
- Custom domain mapping support

### Integration Opportunities
- SAML/OAuth integration for enterprise authentication
- Webhook notifications for document access
- API for programmatic document management
- CDN integration for improved global performance

## Acceptance Criteria

The implementation is considered complete when:

1. ✅ All functional requirements are implemented and tested
2. ✅ Security headers prevent search engine indexing
3. ✅ Standard HTML development workflow works (CSS/JS without tokens)
4. ✅ Successful deployment to Google Cloud Run
5. ✅ CLI tools provide complete document management capabilities
6. ✅ Iframe embedding works in target LMS platforms
7. ✅ Comprehensive documentation covers all use cases
8. ✅ Performance meets specified benchmarks
9. ✅ Security testing validates protection mechanisms
10. ✅ End-to-end testing confirms all workflows

## Conclusion

Cloud Docs successfully balances security requirements with developer productivity by providing a two-tier access model. The platform enables organizations to host private documentation while maintaining the simplicity of standard web development practices, making it ideal for LMS integration and controlled document distribution scenarios.