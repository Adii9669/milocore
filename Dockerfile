# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# RUN go build -o server .
#
# # Final stage
# FROM alpine:latest
#
# WORKDIR /root/
#
# COPY --from=builder /app/server .

EXPOSE 8000

CMD ["air"]
