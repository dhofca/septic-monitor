##
## Build stage (CGO enabled for github.com/mattn/go-sqlite3)
##
FROM golang:1.25-bookworm AS build

WORKDIR /src

# Install toolchain for CGO (sqlite3 driver needs it)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    gcc \
    libc6-dev \
 && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1
RUN go build -trimpath -ldflags="-s -w" -o /out/sceptic-monitor ./main.go

##
## Runtime stage
##
FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
 && rm -rf /var/lib/apt/lists/*

COPY --from=build /out/sceptic-monitor /app/sceptic-monitor

# Fly sets PORT, default code falls back to 8080.
EXPOSE 8080

CMD ["/app/sceptic-monitor"]
