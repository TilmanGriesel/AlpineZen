# Builder stage
FROM golang:1.24.0-alpine as builder

ARG VERSION=latest
ENV APP_VERSION=$VERSION

ARG BUILD_NUM=0
ENV APP_BUILD_NUM=$BUILD_NUM

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o cli ./cmd/cli

# Final stage
FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/cli .
COPY docker/entrypoint.sh /app/entrypoint.sh

RUN apk add --no-cache caddy

RUN mkdir -p /srv/latest_wallpaper

RUN ln -sf /root/.alpinezen_wallpaper/latest.jpg /srv/latest_wallpaper/latest.jpg

RUN chmod +x /app/entrypoint.sh

EXPOSE 80

ENTRYPOINT ["/app/entrypoint.sh"]

CMD []
