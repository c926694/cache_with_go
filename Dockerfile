FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o cache-server ./cmd/cache-server

FROM alpine:3.22

WORKDIR /app

COPY --from=build /app/cache-server .

EXPOSE 3000

CMD ["./cache-server", "-addr", "0.0.0.0:3000", "-capacity", "67108864"]
