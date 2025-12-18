# Household Manager App Server

## Tech Stack

[![Stack](https://skillicons.dev/icons?i=golang,docker,postgresql,redis,gcp,aws,prometheus,grafana&theme=dark&perline=15)]()

## Database Schema

<img width="4080" height="3116" alt="drawSQL-image-export-2025-09-04" src="https://github.com/user-attachments/assets/2fbb9b0f-5b10-4f71-83d2-ca1fd19453a3" />

## API Documentation

This project uses Swagger for API documentation.

- **Swagger UI**: Access the interactive API documentation at `http://localhost:8000/swagger/index.html` when the server is running.
- **Spec Files**: The raw Swagger specification files are located in the `docs/` directory (`swagger.json`, `swagger.yaml`).

## Monitoring & Observability

The application includes comprehensive monitoring with Prometheus and Grafana.

### Available Metrics

| Category | Metrics |
|----------|---------|
| **HTTP** | Request count, duration, in-flight requests, response size |
| **Database** | Query count, duration, connections, errors |
| **Redis** | Operations, cache hits/misses, operation duration |
| **Authentication** | Login attempts, token generation, validation |
| **S3** | Upload count, duration, file size |
| **Business** | Homes, tasks, bills, shopping items, polls |
| **Email** | Sent emails by type and status |
| **OCR** | Request count and processing duration |

### Accessing Monitoring

| Service | URL | Credentials |
|---------|-----|-------------|
| Prometheus | `http://localhost:9090` | - |
| Grafana | `http://localhost:3000` | admin / admin |
| Metrics Endpoint | `http://localhost:8000/metrics` | - |

### Grafana Dashboard

A pre-configured dashboard "Household Manager API" is automatically provisioned with panels for:

- **Overview**: Uptime, request rate, error rate, P95 latency, active users
- **HTTP Metrics**: Request rate by status, latency percentiles, top endpoints
- **Database Metrics**: Queries by operation, query latency, connections
- **Redis Metrics**: Operations by type, cache hit/miss rate
- **Authentication**: Login attempts, token validation
- **Business Metrics**: Homes, tasks, bills, shopping items, polls
- **External Services**: S3 uploads, email sends, OCR requests

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

### Production Grafana Variables (Optional)

| Variable | Description | Default |
|----------|-------------|---------|
| `GRAFANA_ADMIN_USER` | Grafana admin username | `admin` |
| `GRAFANA_ADMIN_PASSWORD` | Grafana admin password | `admin` |
| `GRAFANA_ROOT_URL` | Grafana public URL | `http://localhost:3000` |

## Running the Application

This project uses Docker Compose for easy deployment.

### Development

To run the application in development mode with hot-reloading and monitoring:

```bash
docker compose -f docker-compose.dev.yaml up --build
```

Available services:
- **API Server**: `http://localhost:8000`
- **Swagger UI**: `http://localhost:8000/swagger/index.html`
- **Prometheus**: `http://localhost:9090`
- **Grafana**: `http://localhost:3000`
- **PostgreSQL**: `localhost:5432`
- **Redis**: `localhost:6379`

### Production

To run the application in production mode:

1. Create a `.env.prod` file with your production values.
2. (Optional) Set Grafana credentials in `.env.prod`:
   ```
   GRAFANA_ADMIN_USER=your-admin-user
   GRAFANA_ADMIN_PASSWORD=your-secure-password
   GRAFANA_ROOT_URL=https://grafana.yourdomain.com
   ```
3. Run the production compose file:

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

## Project Structure

```
diploma-server/
├── cmd/server/           # Application entry point
├── internal/
│   ├── cache/            # Redis client
│   ├── config/           # Configuration management
│   ├── http/
│   │   ├── handlers/     # API endpoint handlers
│   │   └── middleware/   # HTTP middleware (metrics, logging, auth)
│   ├── logger/           # Logging utilities
│   ├── metrics/          # Prometheus metrics definitions
│   ├── models/           # Data models
│   ├── repository/       # Data access layer
│   ├── router/           # Route setup
│   ├── services/         # Business logic
│   └── utils/            # Utility functions
├── monitoring/
│   ├── prometheus/       # Prometheus configuration
│   └── grafana/
│       ├── dashboards/   # Grafana dashboard JSON files
│       └── provisioning/ # Grafana auto-provisioning configs
├── docs/                 # Swagger documentation
├── docker-compose.dev.yaml
├── docker-compose.prod.yaml
├── Dockerfile
└── Dockerfile.dev
```
