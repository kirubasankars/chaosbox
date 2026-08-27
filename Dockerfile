# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/chaosbox ./cmd/chaosbox

FROM alpine:3.20
RUN apk add --no-cache ca-certificates \
    && addgroup -S chaosbox \
    && adduser -S -G chaosbox chaosbox \
    && mkdir -p /app/data /app/logs \
    && touch /app/data/file.txt \
    && chown -R chaosbox:chaosbox /app

WORKDIR /app
COPY --from=build /out/chaosbox /usr/local/bin/chaosbox
COPY config.json ./config.json

USER chaosbox
EXPOSE 8080

ENTRYPOINT ["chaosbox"]
CMD []
