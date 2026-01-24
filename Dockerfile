# Stage 1: Build Frontend
FROM node:20-slim AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o echobridge main.go

# Stage 3: Final Image
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates

# Copy binary
COPY --from=backend-builder /app/echobridge .

# Copy frontend build to ./web
COPY --from=frontend-builder /app/frontend/build ./web

# Copy other necessary files (if any, e.g., .env but we are hardcoding/using env vars)
# COPY .env . 

EXPOSE 8000
CMD ["./echobridge"]
