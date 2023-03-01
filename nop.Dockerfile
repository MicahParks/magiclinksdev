FROM golang:1 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o magiclinksdev -trimpath cmd/nop_provider/*.go

FROM alpine
COPY --from=builder /app/magiclinksdev /magiclinksdev
CMD ["/magiclinksdev"]
