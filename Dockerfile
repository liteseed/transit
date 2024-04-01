FROM alpine:latest

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

WORKDIR /transit

VOLUME ["/transit/data"]

COPY cmd/main.go /transit/transit
EXPOSE 8080

ENTRYPOINT [ "cmd/main" ]