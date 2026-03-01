FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY src/get_gcp_token.go .

RUN go build -o get_gcp_token get_gcp_token.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/get_gcp_token .

# Install curl for manual verification if needed
RUN apk add --no-cache curl

# Set env var for testing
ENV AWS_WEB_IDENTITY_TOKEN_FILE=/tmp/token

ENTRYPOINT ["./get_gcp_token"]
