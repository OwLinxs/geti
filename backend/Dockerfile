# ---- Estágio de build ----
FROM golang:1.26-bookworm AS builder

WORKDIR /app

# CGO é necessário para o driver SQLite (mattn/go-sqlite3).
ENV CGO_ENABLED=1

# Dependências vendorizadas (vendor/) — build offline, sem acesso a proxy.golang.org.
COPY . .
RUN go build -mod=vendor -o /sige-ti ./cmd/server

# ---- Estágio final ----
FROM debian:bookworm-slim

# Certificados (para PostgreSQL TLS) e curl (para o HEALTHCHECK).
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /sige-ti /app/sige-ti

# Diretório para o banco SQLite e logo do termo (montados como volume).
RUN mkdir -p /app/data /app/assets

EXPOSE 8080
ENV PORT=8080 \
    DB_DRIVER=sqlite \
    DB_DSN=/app/data/sige-ti.db

# Verificação de saúde: a aplicação expõe /health. start-period cobre a
# migração inicial + seed no primeiro boot.
HEALTHCHECK --interval=15s --timeout=3s --start-period=20s --retries=5 \
    CMD curl -fsS http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/sige-ti"]
