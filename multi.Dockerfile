FROM golang:1 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o magiclinksdev -trimpath cmd/multi_provider/*.go

FROM alpine
COPY --from=builder /app/magiclinksdev /magiclinksdev
CMD ["/magiclinksdev"]
