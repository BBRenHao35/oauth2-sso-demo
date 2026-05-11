FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server .

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8081
CMD ["./server"]
