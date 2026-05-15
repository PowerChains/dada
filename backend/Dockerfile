# Backend Docker image for production
# Build the frontend static assets and the Go backend in one image.
FROM node:20-alpine AS frontend-build
WORKDIR /app
COPY frontend/package.json frontend/vite.config.js ./
COPY frontend/index.html ./
COPY frontend/src ./src
COPY frontend/styles.css ./
RUN npm install
RUN npm run build

FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/*.go ./
COPY --from=frontend-build /app/dist ./static
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/dada .

FROM gcr.io/distroless/cc
WORKDIR /app
COPY --from=builder /app/dada /app/dada
COPY --from=builder /app/static /app/static
EXPOSE 8080
CMD ["/app/dada"]
