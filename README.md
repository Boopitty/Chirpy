# Chirpy
Web server guided by a Boot.dev course.

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) 1.21 or later
- [PostgreSQL](https://www.postgresql.org/download/)

### Install

```bash
# Install the project
git clone github.com/Boopitty/Chirpy.git

# change to the project directory
cd Chirpy

# Download dependancies
go mod download

# copy the .env example file, then update the values with your own local settings.
cp .env.example .env
```



### Configuration

You can generate the secret variable by running this command in your terminal. 

Copy it into the `.env` file once made.
```bash
openssl rand -base64 32
```

Make an up migration to the database.
```bash
goose -dir sql/schema postgres "your_connection_string_here" up
```


### Run it

```bash
# Run the server
go build -o chirpy && ./chirpy
```

## API endpoints

### GET
- `GET /api/healthz`: Health check for the server.

- `GET /admin/metrics`: Gets admin metrics. Currently only the number of times the home page was visited.

- `GET /api/chirps`: Returns a JSON slice of all chirps in the db. If an `author_id` query parameter is provided, it will filter for posts with the given author.

- `GET /api/chirps/{chirpID}`: Gets chirp with specific id in JSON format. Returns status code 400 if chirpID is invalid, and 404 if it does not exist.

### POST
- `POST /admin/reset`: Reset the database.

- `POST /api/chirps`: Add a chirp to the db. Requires an access token.

```json
// Expects JSON
{
    "body": "put the chirp here!"
}

// Responds with the new chirp in JSON
{
    "id": "chirp_uuid",
    "created_at": "2026-05-30T12:00:00Z",
    "updated_at": "2026-05-30T12:00:00Z",
    "body": "User's chirp here",
    "user_id": "user_uuid"
}
```


- `POST /api/users`: Create a new user in the db. 

```json
// Expects JSON
{
    "email": "example@email.com",
    "password": "one234"
}

// Returns the new user information in JSON format
{
    "id": "uuid",
    "created_at": "2026-05-30T12:00:00Z",
    "updated_at": "2026-05-30T12:00:00Z",
    "email": "example@email.com",
    "is_chirpy_red": false
}
```

- `POST /api/login`: Login as an existing user

```json
// Expects JSON
{
    "email": "example@email.com",
    "password": "one234"
}
```

```json
// Responds with JSON
{
    "id": "uuid",
    "created_at": "2026-05-30T12:00:00Z",
    "updated_at": "2026-05-30T12:00:00Z",
    "email": "example@email.com",
    "token": "access_token",
    "refresh_token": "refresh_token",
    "is_chirpy_red": false
}
```

- `POST /api/refresh`: Get a new access token. Requires a refresh token.

```json
// Responds with JSON
{
    "token": "new_token_string"
}
```

- `POST /api/revoke`: Revoke a refresh token. The refresh token to-be-revoked must be in the bearer header. Success returns `204 No Content`

- `POST /api/polka/webhooks`: Handle webhook for polka. Requires the POKLA_KEY from the .env file in the bearer header. Success returns `204 No Content`.

```json
// Expects JSON
{
    "event": "user.upgraded",
    "data": {
        "user_id": "user_uuid"
    }
}
```

### PUT
- `PUT /api/users`: Update a user's email and password in the db.

```json
// Expects JSON
{
    "email": "newemail@example.com",
    "password": "newpass"
}

// Responds with JSON
{
    "id": "user_uuid",
    "created_at": "2026-05-30T12:00:00Z",
    "updated_at": "2026-05-30T12:00:00Z",
    "email": "user@email.com",
    "is_chirpy_red": false
}
```

### DELETE
- `DELETE /api/chirps/{chirpID}`: Delete a chirp with a specifid id. Requires access token.