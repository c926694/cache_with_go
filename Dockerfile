FROM golang:1.26-alpine AS build

WORKDIR /app
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o cache-server ./cmd

FROM alpine:3.22

WORKDIR /app

COPY --from=build /app/cache-server .

EXPOSE 3000

ENTRYPOINT ["./cache-server"]
CMD ["-addr", "0.0.0.0:3000", "-capacity", "67108864"]
