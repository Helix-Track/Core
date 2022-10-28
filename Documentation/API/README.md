# HelixTrack Core REST API documentation

The following sections list all the HeliXTrack API calls specifications.

## Authentication

### Authenticate user

- endpoint: `/action`
- method: `POST`
- payload: 
  ```json
  {
    "action": "authenticate",
    "username": "string",
    "password": "string"
  }
  ```
- response:
  ```json
  {
    "jwt": "string"
  }
  ```


## Core API

Tbd.