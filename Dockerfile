FROM golang:1.22-alpine AS build_env
WORKDIR /app
COPY . .
RUN go build -o deduplicator ./cmd/*.go


FROM alpine:3.20
WORKDIR /app
COPY --from=build_env /app /app
ENTRYPOINT [ "/app/deduplicator" ]
