# Build stage
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git build-base

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /out/warehouse .

# Runtime: git required for pull/push/rm
FROM alpine:3.20
RUN apk add --no-cache ca-certificates git openssh-client

WORKDIR /app
COPY --from=builder /out/warehouse /usr/local/bin/warehouse
RUN chmod +x /usr/local/bin/warehouse
COPY deploy/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["warehouse", "list", "--pull"]
