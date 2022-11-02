# HelixTrack Core REST API documentation

The following sections list all the HeliXTrack API calls specifications.

Table of contents

- [Authentication API](#Authentication-API)
  - [Authenticate the user](#Authenticate-the-user)
    - [The JWT payload](#The-JWT-payload)
  - [Sign out the user from the system](#Sign-out-the-user-from-the-system)
  - [Obtain the version of the authentication API service](#Obtain-the-version-of-the-authentication-API-service)
- [Core API](#Core-API)
  - [JWT capability check](#JWT-capability-check)
  - [Database capability check](#Database-capability-check)
  - [System health](#System-health)
  - [Obtain the version of the HelixTrack Core API service](#Obtain-the-version-of-the-HelixTrack-Core-API-service)
  - [Perform the entity CRUD operations](#Perform-the-entity-CRUD-operations)
- [API response error codes](#API-response-error-codes)
  - [Request related error codes](#Request-related-error-codes)
  - [System related error codes](#System-related-error-codes)
  - [Entity related error codes](#Entity-related-error-codes)
- [HashiCorp Vault API](#HashiCorp-Vault-API)

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

### Obtain the version of the HelixTrack Core API service

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

## [HashiCorp Vault](https://github.com/hashicorp/vault) API

Vault API documentation can be found [here](https://developer.hashicorp.com/vault/api-docs).

*Note:* Vault is used by API services.