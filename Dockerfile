# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=0.1.0
ARG BUILD_DATE=unknown
ARG COMMIT=unknown

RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w -X github.com/mmfpsolutions/gsbe/internal/version.Version=${VERSION} -X github.com/mmfpsolutions/gsbe/internal/version.BuildDate=${BUILD_DATE} -X github.com/mmfpsolutions/gsbe/internal/version.Commit=${COMMIT}" \
    -o gsbe ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates su-exec tzdata

RUN addgroup -g 1000 app && adduser -u 1000 -G app -D app

WORKDIR /app

COPY --from=builder /build/gsbe .
COPY docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

RUN mkdir -p /app/config /app/logs && chown -R app:app /app

EXPOSE 3007

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:3007/health || exit 1

ENTRYPOINT ["./docker-entrypoint.sh"]
