FROM golang:1.14 as builder

WORKDIR /work

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY main.go main.go
COPY cmd/ cmd/
COPY pkg/ pkg/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o torchstand main.go

FROM alpine:3

RUN apk add --no-cache ca-certificates

WORKDIR /

COPY --from=builder /work/torchstand .

ENTRYPOINT ["/torchstand"]
