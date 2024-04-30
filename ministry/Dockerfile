FROM golang:1.22-alpine3.19 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY internal/ internal/
COPY cmd/ cmd/

RUN --mount=type=cache,target="/root/.cache/go-build" GOCACHE=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ministry ./cmd/ministry

FROM alpine:3.19.1

LABEL org.opencontainers.image.created="${BUILD_TIME}"
LABEL org.opencontainers.image.source="https://github.com/glizzus/trf"
LABEL org.opencontainers.image.title="trf-ministry"
LABEL org.opencontainers.image.description="The ministry of truth backend for trf"
LABEL org.opencontainers.image.revision="${GIT_HASH}"

WORKDIR /app
ADD migrations /app/migrations
ADD templates /app/templates
COPY --from=builder /app/ministry .
CMD ["./ministry"]
