# Stage 1: Builder
FROM golang:1.25 AS builder

RUN apt-get update && apt-get install -y \
    libtesseract-dev \
    libleptonica-dev \
    tesseract-ocr \
    tesseract-ocr-rus \
    tesseract-ocr-eng \
    tesseract-ocr-ukr \
    tesseract-ocr-pol \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-w -s" -o main ./cmd/server

# Stage 2: Final image
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    tesseract-ocr-rus \
    tesseract-ocr-eng \
    tesseract-ocr-ukr \
    tesseract-ocr-pol \
    libtesseract5 \
    libleptonica7 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8000
CMD ["./main"]