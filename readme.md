# Chirpy - A Simple Social Media Backend

## Description
Chirpy is a lightweight social media backend service that allows users to post short messages ("chirps"), manage user accounts, and handle authentication. It's built with Go and uses PostgreSQL for data storage.

## Features
- User authentication with JWT tokens and refresh tokens
- Post and retrieve chirps (messages)
- Filter chirps by author
- Sort chirps by creation time
- User account management (create, update, login)
- Premium features (Chirpy Red subscription)
- Profanity filtering for chirps

## Use Cases
Perfect for:
- Learning backend development
- Prototyping social media features
- Building a Twitter-like service
- Understanding JWT authentication

## Prerequisites
- Go 1.21 or later
- PostgreSQL
- [Goose](https://github.com/pressly/goose) for database migrations
- [SQLC](https://sqlc.dev/) for type-safe SQL

## Installation
1. Clone the repository
```bash
git clone https://github.com/yourusername/chirpy.git
cd chirpy
```

2. Setup environment variables in .env:

```bash
DB_URL=postgres://username:password@localhost:5432/dbname?sslmode=disable
JWT_SECRET=your-secret-key
PLATFORM=dev
```

3. Install dependencies

```bash
go mod download
```

4. Run database migrations
```bash
goose -dir sql/migrations postgres "your-db-url" up
```

## Running the project
Start the server:
```bash
go run .
```
The server will start on localhost:8080.

## API Endpoints
POST /api/users - Create new user
POST /api/login - Login user
GET /api/chirps - Get all chirps
POST /api/chirps - Create new chirp
PUT /api/users - Update user details
And more...

*License*
MIT