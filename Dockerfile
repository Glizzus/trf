FROM golang:1.22-alpine3.19 AS builder

# We need ca-certificates for HTTPS requests
RUN apk update && apk add --no-cache ca-certificates=20240226-r0

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY internal/ internal/
COPY cmd/ cmd/
ARG BUILD_CACHE=/root/.cache/go-build
RUN --mount=type=cache,target=${BUILD_CACHE} \
    GOCACHE=${BUILD_CACHE} \
    CGO_ENABLED=0 \
    GOOS=linux \
    go build -ldflags="-s -w" -a -installsuffix cgo -o /app/ministry ./cmd/ministry


FROM scratch AS ministry

ARG BUILD_TIME
LABEL org.opencontainers.image.created="${BUILD_TIME}"

LABEL org.opencontainers.image.source="https://github.com/glizzus/trf"

ARG GIT_HASH
LABEL org.opencontainers.image.revision="${GIT_HASH}"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/ministry /ministry
COPY ./templates /templates
ENTRYPOINT ["/ministry"]
CMD ["serve"]


FROM nginx:1.27.0-alpine3.19-slim AS nginx

COPY ./nginx.conf /etc/nginx/templates/nginx.conf.template
COPY ./static /usr/share/nginx/html
