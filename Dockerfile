FROM golang AS builder

WORKDIR /app
ADD go.mod go.sum ./
RUN go mod download && go mod verify

ADD internal/ internal/
ADD cmd/ cmd/

RUN --mount=type=cache,target="/root/.cache/go-build" GOCACHE=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ministry ./cmd/ministry

FROM alpine:latest
WORKDIR /app
ADD migrations /app/migrations
ADD templates /app/templates
COPY --from=builder /app/ministry .
CMD ["./ministry"]
