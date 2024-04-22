FROM golang AS builder

WORKDIR /app
ADD go.mod go.sum ./
RUN go mod download && go mod verify

ADD internal/ internal/
ADD cmd/ cmd/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ministry ./cmd/ministry

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/ministry .
CMD ["./ministry"]
