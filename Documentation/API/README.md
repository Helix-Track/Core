# HelixTrack Core REST API documentation

The following sections list all the HeliXTrack API calls specifications.

## Authentication API

### Authenticate the user

- endpoint: `/do`
- method: `POST`
- payload: 
  ```yaml
  {
    "action": "authenticate", /* mandatory */
    "username": "string",     /* mandatory */
    "password": "string",     /* mandatory */
    "locale": "string?"       /* optional  */
  }
  ```
- response:
  ```yaml
  {
    "jwt": "string",
    "errorCode": -1,
    "errorMessage": "string",
    "errorMessageLocalised": "string"
  }
  ```


## Core API

Tbd.