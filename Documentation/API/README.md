# helix track Core REST API documentation

The following sections list all the helix track API calls specifications.

Table of contents

- [helix track Core REST API documentation](#helix-track-core-rest-api-documentation)
  - [Authentication API](#authentication-api)
    - [Authenticate the user](#authenticate-the-user)
      - [The JWT payload](#the-jwt-payload)
    - [Sign out the user from the system](#sign-out-the-user-from-the-system)
    - [Obtain the version of the authentication API service](#obtain-the-version-of-the-authentication-api-service)
  - [Core API](#core-api)
    - [JWT capability check](#jwt-capability-check)
    - [Database capability check](#database-capability-check)
    - [System health](#system-health)
    - [Obtain the version of the helix track Core API service](#obtain-the-version-of-the-helix-track-core-api-service)
    - [Perform the entity CRUD operations](#perform-the-entity-crud-operations)
  - [API response error codes](#api-response-error-codes)
    - [Request related error codes](#request-related-error-codes)
    - [System related error codes](#system-related-error-codes)
    - [Entity related error codes](#entity-related-error-codes)
  - [HashiCorp Vault API](#hashicorp-vault-api)

## Authentication API

### Authenticate the user

- endpoint: `/do`
- method: `POST`
- payload:
  ```yaml
  {
    "action":     "authenticate"   /* mandatory */
    "username":   "string",        /* mandatory */
    "password":   "string",        /* mandatory */
    "locale":     "string"         /* optional  */
  }
  ```
- response:
  ```yaml
  {
    "jwt":                    "string",
    "errorCode":              -1,
    "errorMessage":           "string",
    "errorMessageLocalised":  "string"
  }
  ```
  
#### The JWT payload

```yaml
{
  "sub":                       "authentication",
  "name":                      "string",
  "username":                  "string",
  "role":                      "string",
  "permissions":               "string",
  "htCoreAddress":             "string"
}
```

### Sign out the user from the system

- endpoint: `/do`
- method: `POST`
- payload:
  ```yaml
  {
    "action":      "signOut"      /* mandatory */
    "jwt":         "string"       /* mandatory */
    "locale":      "string"       /* optional  */
  }
  ```
- response:
  ```yaml
  {
    "errorCode":              -1,
    "errorMessage":           "string",
    "errorMessageLocalised":  "string"
  }
  ```

### Obtain the version of the authentication API service

- endpoint: `/do`
- method: `POST`
- payload:
  ```yaml
  {
    "action":      "version"      /* mandatory */
    "jwt":         "string"       /* mandatory */
    "locale":      "string"       /* optional  */
  }
  ```
- response:
  ```yaml
  {
    "version":                "string",
    "versionCode":            "string",
    "errorCode":              -1,
    "errorMessage":           "string",
    "errorMessageLocalised":  "string"
  }
  ```

## Core API

### JWT capability check

- endpoint: `/do`
- method: `POST`
- payload:
  ```yaml
  {
    "action":      "jwtCapable"   /* mandatory */
    "locale":      "string"       /* optional  */
  }
  ```
- response:
  ```yaml
  {
    "capable":                "boolean",
    "errorCode":              -1,
    "errorMessage":           "string",
    "errorMessageLocalised":  "string"
  }
  ```

### Database capability check

- endpoint: `/do`
- method: `POST`
- payload:
  ```yaml
  {
    "action":      "dbCapable"    /* mandatory */
    "locale":      "string"       /* optional  */
  }
  ```
- response:
  ```yaml
  {
    "capable":                "boolean",
    "errorCode":              -1,
    "errorMessage":           "string",
    "errorMessageLocalised":  "string"
  }
  ```

### System health

- endpoint: `/do`
- method: `POST`
- payload:
  ```yaml
  {
    "action":      "health"       /* mandatory */
    "locale":      "string"       /* optional  */
  }
  ```
- response:
  ```yaml
  {
    "report":                 "string",
    "errorCode":              -1,
    "errorMessage":           "string",
    "errorMessageLocalised":  "string"
  }
  ```

### Obtain the version of the helix track Core API service

- endpoint: `/do`
- method: `POST`
- payload:
  ```yaml
  {
    "action":      "version"      /* mandatory */
    "jwt":         "string"       /* mandatory */
    "locale":      "string"       /* optional  */
  }
  ```
- response:
  ```yaml
  {
    "version":                "string",
    "versionCode":            "string",
    "errorCode":              -1,
    "errorMessage":           "string",
    "errorMessageLocalised":  "string"
  }
  ```

### Perform the entity CRUD operations

- endpoint: `/do`
- method: `POST`
- payload:
  ```yaml
  {
    "action":      "string"       /* mandatory */
    "jwt":         "string"       /* mandatory */
    "locale":      "string"       /* optional  */
    "object":      "string"       /* mandatory */
  }
  ```
- response:
  ```yaml
  {
    "id":                     "string",
    "errorCode":              -1,
    "errorMessage":           "string",
    "errorMessageLocalised":  "string"
  }
  ```
  
Supported `action`(s) are:

- `create`
- `modify`
- `remove`.

The `object` payload can be:

- JSON serialized value of the object if `action` is `create` or `modify`
- The UUID identifier of the object if it is the `remove` `action`.

Response will contain the UUID identifier of the object with no error code (-1) 
if operation is successful.

## API response error codes

The following list contains all supported error codes:

- -1:   No error
- 100X: Request related error codes
- 200X: System related error codes
- 300X: Entity related error codes

*Note:* More error codes will be documented soon.

### Request related error codes

Tbd.

### System related error codes

Tbd.

### Entity related error codes

Tbd.

## REST API Endpoints

### Authentication Endpoints

#### Register User
- **Endpoint**: `POST /api/auth/register`
- **Description**: Register a new user account
- **Request Body**:
  ```json
  {
    "username": "string",
    "password": "string",
    "email": "string",
    "name": "string"
  }
  ```
- **Response**:
  ```json
  {
    "id": "string",
    "username": "string",
    "email": "string",
    "name": "string",
    "role": "string"
  }
  ```

#### Login User
- **Endpoint**: `POST /api/auth/login`
- **Description**: Authenticate user and get JWT token
- **Request Body**:
  ```json
  {
    "username": "string",
    "password": "string"
  }
  ```
- **Response**:
  ```json
  {
    "token": "string",
    "username": "string",
    "email": "string",
    "name": "string",
    "role": "string"
  }
  ```

#### Logout User
- **Endpoint**: `POST /api/auth/logout`
- **Description**: Logout user (stateless JWT system)
- **Response**:
  ```json
  {
    "message": "Successfully logged out"
  }
  ```

### Service Discovery Endpoints

#### Register Service
- **Endpoint**: `POST /api/services/register`
- **Description**: Register a new service in the discovery system
- **Authentication**: Required (Admin)

#### Discover Services
- **Endpoint**: `POST /api/services/discover`
- **Description**: Discover available services
- **Authentication**: Required (Admin)

#### Rotate Service
- **Endpoint**: `POST /api/services/rotate`
- **Description**: Rotate service instances
- **Authentication**: Required (Admin)

#### Decommission Service
- **Endpoint**: `POST /api/services/decommission`
- **Description**: Remove a service from discovery
- **Authentication**: Required (Admin)

#### Update Service
- **Endpoint**: `POST /api/services/update`
- **Description**: Update service information
- **Authentication**: Required (Admin)

#### List Services
- **Endpoint**: `GET /api/services/list`
- **Description**: Get list of all registered services
- **Authentication**: Required (Admin)

#### Get Service Health
- **Endpoint**: `GET /api/services/health/:id`
- **Description**: Get health status of a specific service
- **Authentication**: Required (Admin)

### WebSocket Endpoints (if enabled)

#### WebSocket Connection
- **Endpoint**: `GET /ws` (configurable path)
- **Description**: Establish WebSocket connection for real-time events

#### WebSocket Stats
- **Endpoint**: `GET /ws/stats`
- **Description**: Get WebSocket connection statistics

### System Endpoints

#### Health Check
- **Endpoint**: `GET /health`
- **Description**: Check system health status
- **Response**:
  ```json
  {
    "status": "ok"
  }
  ```

## HashiCorp Vault API

[HashiCorp Vault](https://github.com/hashicorp/vault) API documentation can be found [here](https://developer.hashicorp.com/vault/api-docs).

*Note:* Vault is used by the API services.