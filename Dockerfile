# ============================================
# Stage 1: Builder
# ============================================
FROM golang:1.23-alpine AS builder

# Instalar dependencias necesarias para compilar
RUN apk add --no-cache git gcc musl-dev

# Crear directorio de trabajo
WORKDIR /app

# Copiar go.mod y go.sum primero (para cache de dependencias)
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
# CGO_ENABLED=1 porque valkey-glide usa CGO
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# ============================================
# Stage 2: Runtime
# ============================================
FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder /app/server .

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 50051

# Variables de entorno con defaults para desarrollo
ENV VALKEY_HOST=localhost \
    VALKEY_PORT=6379 \
    VALKEY_PASSWORD="" \
    VALKEY_IS_CLUSTER=false

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD nc -z localhost 50051 || exit 1

CMD ["./server"]