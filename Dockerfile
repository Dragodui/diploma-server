# Stage 1: Builder
FROM golang:1.25 AS builder

# Устанавливаем Tesseract и dev-библиотеки
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

# Ставим зависимости Go
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник (CGO включён по умолчанию)
RUN go build -ldflags="-w -s" -o main ./cmd/server

# Stage 2: Final image
FROM debian:bullseye-slim

# Устанавливаем Tesseract для runtime
RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    tesseract-ocr-rus \
    tesseract-ocr-eng \
    tesseract-ocr-ukr \
    tesseract-ocr-pol \
    libtesseract-dev \
    libleptonica-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/main .

# Expose port
EXPOSE 8000

# Run the binary
CMD ["./main"]
