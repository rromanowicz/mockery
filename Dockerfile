FROM golang:1.25.6-alpine AS builder

RUN apk update \
    && apk add --no-cache git \
    && apk add --update gcc alpine-sdk

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o router .
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/router .
EXPOSE 8080
CMD ["./router"]

# docker build -t router-app .
# docker run --name "router-app" -p 8080:8080 router-app
