# HelixTrack Core REST API documentation

The following sections list all the HeliXTrack API calls specifications.

## Authentication

### Authenticate user

- endpoint: `/do`
- method: `POST`
- payload: 
  ```json
  {
    "action": "authenticate",
    "username": "string",
    "password": "string",
    "locale": "string"
  }
  ```
- response:
  ```json
  {
    "jwt": "string",
    "errorCode": -1,
    "errorMessage": "string",
    "errorMessageLocalised": "string"
  }
  ```


## Core API

Tbd.