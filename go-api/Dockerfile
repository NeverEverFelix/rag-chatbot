# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY main.go ./
RUN go mod init ragserver && go mod tidy
RUN go build -o rag-api

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/rag-api ./

EXPOSE 8080
ENTRYPOINT ["./rag-api"]