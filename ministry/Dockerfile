FROM golang:1.22-alpine3.19 AS builder

RUN apk update && apk add --no-cache ca-certificates=20240226-r0

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY internal/ internal/
COPY cmd/ cmd/

RUN --mount=type=cache,target="/root/.cache/go-build" \
    GOCACHE=/root/.cache/go-build CGO_ENABLED=0 \
    GOOS=linux \
    go build -ldflags="-s -w" -a -installsuffix cgo -o ministry ./cmd/ministry

# I have no qualms moving to alpine or something in the future, but
# scratch works for now and speeds up the build process substantially.
FROM scratch

LABEL org.opencontainers.image.created="${BUILD_TIME}"
LABEL org.opencontainers.image.source="https://github.com/glizzus/trf"
LABEL org.opencontainers.image.title="trf-ministry"
LABEL org.opencontainers.image.description="The ministry of truth backend for trf"
LABEL org.opencontainers.image.revision="${GIT_HASH}"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY migrations /migrations
COPY templates /templates
COPY --from=builder /app/ministry /
CMD ["./ministry"]
