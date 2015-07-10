# Auth service

## Task

Service's role is to authenticate user and provide information about the user (email, name, avatar).

## How it works

Service uses providers like Google, Facebook, Twitter to authenticate user.

Once authenticated user object, returned by provider, is stored in permanent storage. User is then recognised based on web cookie.

## Endpoints

| URI                         | Method | Name                |
|-----------------------------|--------|---------------------|
| /who/                       | GET    | Current user        |
| /providers/                 | GET    | Auth providers      |
| /login/{provider}/          | GET    | Login with provider |
| /logout/                    | GET    | Logout              |

### Current user

URI: `/who/`

Method: `GET`

Returns: Current user object or `401 Unauthorized` when user is not logged in
Example:

`GET /who/`

```json
{
    "name": "John Doe",
    "email": "jd@example.com",
    "admin": "true",
    "avatar": "https://www.example.com/img/jd.jpg"
}
```

### Auth providers

URI: `/providers/`

Method: `GET`

Returns: Map of allowed providers and URI to use to login in with them
Example:

`GET /providers/`

```json
{
    "gplus": "/login/gplus/",
    "facebook": "/login/facebook/"
}
```

### Login with provider

URI: `/login/{provider}/`

Method: `GET`

Returns: Redirection to provider's authentication page

### Logout

URI: `/logout/`

Method: `GET`

Returns: Log user out of the system
