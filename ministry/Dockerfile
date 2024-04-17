FROM golang:1.22 AS build

COPY main.go go.mod go.sum /app/
WORKDIR /app

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ministry .

FROM scratch AS runtime

COPY --from=build /app/ministry /ministry

ENTRYPOINT ["/ministry"]
