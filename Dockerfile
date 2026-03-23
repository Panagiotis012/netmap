# Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Build backend
FROM golang:1.22-alpine AS backend
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./cmd/netmap/dist
RUN CGO_ENABLED=1 go build -ldflags "-s -w" -o netmap ./cmd/netmap

# Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=backend /app/netmap /usr/local/bin/netmap
EXPOSE 8080
ENTRYPOINT ["netmap"]
