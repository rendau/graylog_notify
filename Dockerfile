FROM alpine:latest

RUN apk add --no-cache --upgrade ca-certificates tzdata curl

WORKDIR /app

COPY ./cmd/build/. ./
COPY ./docs ./docs

HEALTHCHECK --start-period=5s --interval=10s --timeout=2s --retries=10 CMD curl -f http://localhost/healthcheck || false

CMD ["./svc"]
