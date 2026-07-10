FROM golang:1.24-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app \
    && mkdir -p /data/uploads \
    && chown -R app:app /data
USER app
COPY --from=build /out/api /usr/local/bin/rentnest-api
EXPOSE 8081
ENTRYPOINT ["rentnest-api"]
