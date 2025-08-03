FROM golang:1.24-alpine3.22 AS builder

RUN apk add --no-cache gcc musl-dev pkgconfig sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 GOOS=linux \
  go build -a -ldflags '-linkmode external -extldflags "-static"' -o /app/main .

FROM scratch

COPY --from=builder /app/main .

EXPOSE 8080
ENTRYPOINT ["/main"]
