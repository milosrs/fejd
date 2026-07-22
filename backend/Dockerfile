FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .
RUN swag init --parseDependency --parseInternal -g main.go -o docs
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/backend ./main.go

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/backend /app/backend
COPY --from=builder /src/migrations /app/migrations
COPY --from=builder /src/docs /app/docs
EXPOSE 8080
CMD ["/app/backend"]
