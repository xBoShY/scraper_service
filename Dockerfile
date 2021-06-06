FROM golang:alpine as builder

WORKDIR /app
COPY ./*.go ./go.* ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o start -ldflags="-w -s" .

FROM scratch

WORKDIR /app
COPY --from=builder /app/start /app/

ENTRYPOINT ["/app/start"]
