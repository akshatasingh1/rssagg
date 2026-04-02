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

- Go 1.26+
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

3. Set up the database:
   - Create a PostgreSQL database
   - Run the migrations:
     ```bash
     sqlc generate
     ```

4. Create a `.env` file in the root directory:
   ```
   PORT=8080
   DB_URL=postgres://username:password@localhost:5432/rssagg?sslmode=disable
   ```

5. Start Redis server (if using Docker):
   ```bash
   docker run -d -p 6379:6379 redis:alpine
   ```

### Frontend Setup

1. Navigate to the client directory:
   ```bash
   cd client
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start the development server:
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