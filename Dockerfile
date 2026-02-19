FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY app/go.mod app/go.sum ./
RUN go mod download

COPY app/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

COPY --from=builder /app /usr/local/bin/app

USER appuser

EXPOSE 8080

ENTRYPOINT ["app"]
