# HelixTrack Core REST API documentation

The following sections list all the HeliXTrack API calls specifications.

Table of contents

- [Authentication API](#Authentication-API)
  - [Authenticate the user](#Authenticate-the-user)
  - [Sign out the user from the system](#Sign-out-the-user-from-the-system)
  - [Obtain the version of the authentication API service](#Obtain-the-version-of-the-authentication-API-service)

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
    "jwtCapable":             "boolean",
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