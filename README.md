# jackstand-api
API for managing passwords.

### Examples

#### Signing in
`POST /users/signin`
```json
{
    "email":"test@test.com",
    "password":"test123",
    "returnSecureToken": true
}
```

#### Getting credentials
`GET /users/credentials`

#### Create credential
`POST /users/credentials`

Only `service`, `username` and `password` fields are required.

```json
{
    "service":"test service",
    "username":"username", 
    "password":"password",
    "description": "description of credential",
    "metadata": {"jsonMeta": "data here"}
}
```
