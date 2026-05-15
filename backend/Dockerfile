# Backend Docker image for production
FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/*.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/dada .

FROM gcr.io/distroless/cc
WORKDIR /app
COPY --from=builder /app/dada /app/dada
EXPOSE 8080
CMD ["/app/dada"]
