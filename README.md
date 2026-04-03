# RSS Aggregator

A web application for aggregating and managing RSS feeds, built with a Go backend and React frontend.

## Features

- User authentication and management
- Add and follow RSS feeds
- Automatic feed scraping and post aggregation
- AI-powered post summaries using Gemini API
- Search functionality for posts
- RESTful API for feed operations
- Modern React UI with Tailwind CSS

## Tech Stack

- **Backend**: Go, Chi router, PostgreSQL, Redis
- **Frontend**: React, Vite, Tailwind CSS
- **Database**: PostgreSQL with sqlc for type-safe queries
- **Caching**: Redis for performance

## Installation

### Prerequisites

- Go 1.26.1+
- Node.js 18+
- PostgreSQL
- Redis

### Backend Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/akshatasingh1/rssagg.git
   cd rssagg
   ```

2. Install Go dependencies:
   ```bash
   go mod download
   ```

3. Generate database code:
   ```bash
   sqlc generate
   ```

4. Set up the database:
   - Create a PostgreSQL database named `rssagg`
   - Apply the database schema. For production-level setup, consider using a migration tool like [Goose](https://github.com/pressly/goose) or [golang-migrate](https://github.com/golang-migrate/migrate). If using Goose:
     ```bash
     go install github.com/pressly/goose/v3/cmd/goose@latest
     cd sql/schema
     goose postgres "postgres://username:password@localhost:5432/rssagg?sslmode=disable" up
     ```
     Otherwise, run the SQL files manually in order (001_users.sql through 009_feed_http_headers.sql). For example, using psql:
     ```bash
     psql -U username -d rssagg -f sql/schema/001_users.sql
     # Repeat for each file
     ```

5. Create a `.env` file in the root directory (replace `username` and `password` with your actual PostgreSQL credentials):
   ```
   PORT=8080
   DB_URL=postgres://username:password@localhost:5432/rssagg?sslmode=disable
   REDIS_URL=redis://localhost:6379
   GEMINI_API_KEY=your_gemini_api_key_here
   ```

6. Start PostgreSQL and Redis servers (if not using Docker Compose):
   ```bash
   docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_USER=username -e POSTGRES_DB=rssagg postgres:alpine
   docker run -d -p 6379:6379 redis:alpine
   ```

### Frontend Setup

1. Navigate to the client directory:
   ```bash
   cd client
   ```

2. Create a `.env.local` file to configure the API base URL:
   ```
   VITE_API_BASE_URL=http://localhost:8080/v1
   ```

3. Install dependencies:
   ```bash
   npm install
   ```

4. Start the development server:
   ```bash
   npm run dev
   ```

## Usage

1. Start the backend server:
   ```bash
   go run main.go
   ```

2. The API will be available at `http://localhost:8080`

3. The frontend will be available at `http://localhost:5173` (default Vite port)

## Docker

The project includes a multi-stage `Dockerfile` and `docker-compose.yml` for local development:

- Builds Go app in `golang:1.26-alpine`
- Copies compiled binary into tiny `alpine:latest`
- Exposes port `8080`
- Launches PostgreSQL in `db` service and Redis in `redis` service

### Run with Docker Compose

From project root:

```bash
# build and start services
docker compose up --build -d

# check logs
docker compose logs -f api
```

### Environment

The Compose service sets:

- `PORT=8080`
- `DB_URL=postgres://username:password@db:5432/rssagg?sslmode=disable`
- `REDIS_URL=redis://redis:6379`

If you want local non-Docker run, set `.env` accordingly:

```env
PORT=8080
DB_URL=postgres://username:password@localhost:5432/rssagg?sslmode=disable
REDIS_URL=redis://localhost:6379
```

## API Endpoints

- `GET /v1/healthz` - Health check
- `POST /v1/users` - Create user
- `GET /v1/feeds` - Get feeds
- `POST /v1/feeds` - Create feed
- And more... (see handlers for full list)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.