FROM golang:1.23.3-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o kvstored ./main.go


FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/kvstored .
COPY .env.example .env
EXPOSE 8000

CMD ["./kvstored"]