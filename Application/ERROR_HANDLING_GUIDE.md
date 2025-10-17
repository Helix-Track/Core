# Error Handling Guide

## Overview

This document provides comprehensive information about error handling across all HelixTrack client applications and the Core API.

## Core API Error Structure

The Core API uses a structured error response format that provides consistent error handling across all clients.

### Error Response Format

```json
{
  "errorCode": 1008,
  "errorMessage": "Unauthorized",
  "errorMessageLocalised": "You do not have permission to access this resource",
  "data": null
}
```

### Error Codes

#### Request-Related Errors (100X)
- `1000` - Invalid request
- `1001` - Invalid action
- `1002` - Missing JWT token
- `1003` - Invalid JWT token
- `1004` - Missing object type
- `1005` - Invalid object type
- `1006` - Missing required data
- `1007` - Invalid data format
- `1008` - Unauthorized
- `1009` - Forbidden - insufficient permissions

#### System-Related Errors (200X)
- `2000` - Internal server error
- `2001` - Database error
- `2002` - Service unavailable
- `2003` - Configuration error
- `2004` - Authentication service error
- `2005` - Permission service error
- `2006` - Extension service error

#### Entity-Related Errors (300X)
- `3000` - Entity not found
- `3001` - Entity already exists
- `3002` - Entity validation failed
- `3003` - Failed to delete entity
- `3004` - Failed to update entity
- `3005` - Failed to create entity
- `3006` - Version conflict

## Client Applications

### Web Client (Angular)

#### Error Interceptor
The Web Client uses an HTTP interceptor that:
- Maps Core API error codes to translation keys
- Emits error events for components to handle
- Provides retry logic for retryable errors
- Shows user-friendly error messages

#### Error Event System
Components can subscribe to error events:
```typescript
this.errorInterceptor.onError((error) => {
  if (error.retryable) {
    // Show retry button
  }
  if (error.requiresAuth) {
    // Redirect to login
  }
});
```

### Desktop Client (Tauri + Angular)

#### Enhanced Error Handling
Similar to Web Client but with desktop-specific features:
- Toast notifications with different styles based on error type
- Longer timeout for retryable errors
- Platform-specific error recovery options

### Android Client (Kotlin)

#### Structured Error Handling
- `ErrorHandler` utility class for consistent error processing
- `BaseRepository` class for common error handling patterns
- `BaseViewModel` class for UI error handling

#### Error Handler Usage
```kotlin
val errorMessage = ErrorHandler.getErrorMessage(context, errorCode, fallbackMessage)
val isRetryable = ErrorHandler.isRetryableError(errorCode)
val requiresAuth = ErrorHandler.requiresAuthentication(errorCode)
```

## Error Recovery Strategies

### Retryable Errors
Errors that can be retried automatically or with user intervention:
- `2002` - Service unavailable
- `2003` - Configuration error
- `2004` - Authentication service error
- `2005` - Permission service error
- `2006` - Extension service error

### Authentication Required Errors
Errors that require user authentication:
- `1002` - Missing JWT token
- `1003` - Invalid JWT token
- `1008` - Unauthorized

### Validation Errors
Errors that require user input correction:
- `1006` - Missing required data
- `1007` - Invalid data format
- `3002` - Entity validation failed

## Localization

### Supported Locales
- `en` - English (default)
- `es` - Spanish
- `fr` - French
- `de` - German
- `ja` - Japanese

### Error Message Localization
All error messages are localized using the client's locale preference. The Core API returns both English (`errorMessage`) and localized (`errorMessageLocalised`) versions.

## Testing

### Unit Tests
- Error handler utility functions
- Repository error handling
- ViewModel error processing
- Component error display

### Integration Tests
- API error response handling
- Network error scenarios
- Authentication error flows
- Data validation errors

### End-to-End Tests
- Complete error scenarios across all clients
- Error recovery workflows
- User experience validation

### AI QA Tests
- Automated testing with real browsers/emulators
- Comprehensive error scenario coverage
- User experience validation

## Best Practices

### For Developers
1. Always use structured error codes from the Core API
2. Never hardcode error messages in client applications
3. Provide clear user feedback for all error scenarios
4. Implement appropriate retry logic for retryable errors
5. Handle authentication errors by redirecting to login

### For Users
1. Clear error messages explain what went wrong
2. Actionable guidance on how to resolve issues
3. Retry options for temporary problems
4. Seamless authentication flows

## Troubleshooting

### Common Issues
1. **Network Errors**: Check connectivity and retry
2. **Authentication Errors**: Verify credentials and re-login
3. **Permission Errors**: Contact administrator for access
4. **Validation Errors**: Check input data and correct
5. **Server Errors**: Wait and retry, or contact support

### Debug Information
All error responses include detailed information for debugging:
- Error code for programmatic handling
- Localized message for user display
- Original error message for technical debugging
- Request context for troubleshooting

## Maintenance

### Adding New Error Codes
1. Add error code to `models/errors.go`
2. Add error message to all client localization files
3. Update error handling logic in all clients
4. Add tests for new error scenarios

### Updating Error Messages
1. Update messages in `models/errors.go`
2. Update localization files in all clients
3. Test changes across all platforms
4. Deploy updates to all clients

## Related Documentation

- [API Reference](../Documentation/API_REFERENCE.md)
- [Client Development Guide](../Documentation/CLIENT_DEV_GUIDE.md)
- [Testing Guide](../Documentation/TESTING_GUIDE.md)
- [Localization Guide](../Documentation/LOCALIZATION_GUIDE.md)