# Gateway API Examples

The gateway proxies requests to the auth service. All auth endpoints are prefixed with `/v1/auth/`.

## Register User

Create a new user account. Requires admin authentication. The role is optional and defaults to "viewer".

**Default Admin User:**  
Username: `admin`  
Password: `tempadmin123`  
*Change this password after first login!*

```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "securepassword",
    "role": "viewer"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Login User

Authenticate an existing user.

```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "securepassword"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Change Password

Change the password for the authenticated user.

```bash
curl -X PUT http://localhost:8080/v1/auth/change-password \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "password": "newsecurepassword"
  }'
```

Response: 204 No Content

## List Users

Get a list of all users. Requires admin authentication.

```bash
curl -X GET http://localhost:8080/v1/users \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN"
```

Response:
```json
[
  {
    "id": 1,
    "username": "testuser",
    "role": "viewer",
    "created_at": "2023-01-01T00:00:00Z"
  }
]
```

## Get User

Get a specific user by ID. Requires admin authentication.

```bash
curl -X GET http://localhost:8080/v1/users/1 \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN"
```

Response:
```json
{
  "id": 1,
  "username": "testuser",
  "role": "viewer",
  "created_at": "2023-01-01T00:00:00Z"
}
```

## Update User

Update a user's information. Requires admin authentication.

```bash
curl -X PUT http://localhost:8080/v1/users/1 \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "updateduser",
    "password": "newpassword",
    "role": "editor"
  }'
```

Response:
```json
{
  "id": 1,
  "username": "updateduser",
  "role": "editor",
  "created_at": "2023-01-01T00:00:00Z"
}
```

## Delete User

Delete a user by ID. Requires admin authentication.

```bash
curl -X DELETE http://localhost:8080/v1/users/1 \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN"
```

Response: 204 No Content

## Create Role

Create a new role. Requires admin authentication.

```bash
curl -X POST http://localhost:8080/v1/roles \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "moderator",
    "description": "Can moderate content"
  }'
```

Response:
```json
{
  "id": 4,
  "name": "moderator",
  "description": "Can moderate content",
  "created_at": "2023-01-01T00:00:00Z"
}
```

## List Roles

Get a list of all roles. Requires admin authentication.

```bash
curl -X GET http://localhost:8080/v1/roles \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN"
```

Response:
```json
[
  {
    "id": 1,
    "name": "admin",
    "description": "Full access",
    "created_at": "2023-01-01T00:00:00Z"
  },
  {
    "id": 2,
    "name": "editor",
    "description": "Can edit content",
    "created_at": "2023-01-01T00:00:00Z"
  },
  {
    "id": 3,
    "name": "viewer",
    "description": "Read-only access",
    "created_at": "2023-01-01T00:00:00Z"
  }
]
```

## Get Role

Get a specific role by ID. Requires admin authentication.

```bash
curl -X GET http://localhost:8080/v1/roles/1 \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN"
```

Response:
```json
{
  "id": 1,
  "name": "admin",
  "description": "Full access",
  "created_at": "2023-01-01T00:00:00Z"
}
```

## Update Role

Update a role's information. Requires admin authentication.

```bash
curl -X PUT http://localhost:8080/v1/roles/1 \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "superadmin",
    "description": "Ultimate access"
  }'
```

Response:
```json
{
  "id": 1,
  "name": "superadmin",
  "description": "Ultimate access",
  "created_at": "2023-01-01T00:00:00Z"
}
```

## Delete Role

Delete a role by ID. Requires admin authentication.

```bash
curl -X DELETE http://localhost:8080/v1/roles/4 \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN"
```

Response: 204 No Content