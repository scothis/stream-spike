FROM golang:1.10 as builder

WORKDIR /go/src/github.com/scothis/stream-spike

COPY cmd/ cmd/
COPY pkg/ pkg/
COPY vendor/ vendor/

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o /main cmd/main.go

###########

FROM scratch

# The following line forces the creation of a /tmp directory
WORKDIR /tmp

WORKDIR /

COPY --from=builder /main /main

ENTRYPOINT ["/main"]
