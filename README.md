# Household manager app server

## Tech stack
[![Stach](https://skillicons.dev/icons?i=golang,docker,postgresql,redis,gcp,aws&theme=dark&perline=15)]()

## Database schema
<img width="4080" height="3116" alt="drawSQL-image-export-2025-09-04" src="https://github.com/user-attachments/assets/2fbb9b0f-5b10-4f71-83d2-ca1fd19453a3" />

## API Documentation

This project uses Swagger for API documentation.

- **Swagger UI**: Access the interactive API documentation at `http://localhost:8000/swagger/index.html` when the server is running.
- **Spec Files**: The raw Swagger specification files are located in the `docs/` directory (`swagger.json`, `swagger.yaml`).

## Configuration

The application is configured using environment variables. You can copy the example file to get started:

```bash
cp .env.example .env.dev
```

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_DSN` | PostgreSQL connection string | `postgres://postgres:postgres@localhost:5432/db` |
| `PORT` | Server port | `8000` |
| `JWT_SECRET` | Secret key for JWT signing | `your-secret-key` |
| `CLIENT_URL` | Frontend application URL (for CORS) | `http://localhost:8081` |
| `GOOGLE_CLIENT_ID` | Google OAuth Client ID | `your-google-client-id` |
| `GOOGLE_CLIENT_SECRET` | Google OAuth Client Secret | `your-google-client-secret` |
| `CLIENT_CALLBACK_URL` | OAuth callback URL | `http://localhost:8000/auth/google/callback` |
| `REDIS_ADDR` | Redis address | `redis:6379` |
| `REDIS_PASSWORD` | Redis password | `your-redis-password` |
| `SMTP_HOST` | SMTP server host | `smtp.example.com` |
| `SMTP_PORT` | SMTP server port | `465` |
| `SMTP_USER` | SMTP username | `user@example.com` |
| `SMTP_PASSWORD` | SMTP password | `your-smtp-password` |
| `SMTP_FROM` | Email sender address | `no-reply@example.com` |
| `AWS_ACCESS_KEY` | AWS Access Key ID | `your-aws-access-key` |
| `AWS_SECRET_ACCESS_KEY` | AWS Secret Access Key | `your-aws-secret-key` |
| `AWS_REGION` | AWS Region | `us-east-1` |
| `AWS_S3_BUCKET` | AWS S3 Bucket name | `your-s3-bucket` |

## Running the Application

This project uses Docker Compose for easy deployment.

### Development

To run the application in development mode with hot-reloading (if configured) and exposed ports:

```bash
docker compose -f docker-compose.dev.yaml up --build
```

The server will be available at `http://localhost:8000`.

### Production

To run the application in production mode:

1. Create a `.env.prod` file with your production values.
2. Run the production compose file:

```bash
docker compose -f docker-compose.prod.yaml up --build -d
```

### Local Development (without Docker)

If you prefer to run the Go server locally without Docker (requires a running Postgres and Redis instance):

1. Ensure Postgres and Redis are running and accessible.
2. Update `.env` with the correct `DB_DSN` and `REDIS_ADDR` (e.g., `localhost`).
3. Run the server:

```bash
go run cmd/server/main.go
```
