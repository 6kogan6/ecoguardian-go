FROM golang:1.25.3-bookworm AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends gcc libc6-dev && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY web ./web

RUN CGO_ENABLED=1 GOOS=linux go build -o ecoguardian ./cmd/server

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates libsqlite3-0 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/ecoguardian /app/ecoguardian
COPY --from=builder /app/web /app/web

RUN mkdir -p /data

ENV PORT=8080
ENV DATABASE_PATH=/data/ecoguardian.db
ENV SIMULATION_INTERVAL_SECONDS=5

EXPOSE 8080

CMD ["/app/ecoguardian"]
