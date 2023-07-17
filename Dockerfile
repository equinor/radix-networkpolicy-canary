# Application build stage
FROM golang:1.19-alpine3.17 as build

ENV GOPATH /go

COPY . /go/src/github.com/equinor/radix-networkpolicy-canary/

WORKDIR /go/src/github.com/equinor/radix-networkpolicy-canary/

RUN apk add --no-cache git
RUN go get -t -v ./... # go get -d -v
RUN go build -v cmd/main.go

# Application run stage
FROM alpine:3

# Add bash if you need an interactive shell in the container, adds ~4MB to final image
# RUN apk add --no-cache bash

RUN addgroup user && adduser -D -G user 2000
USER 2000

COPY --from=build /go/src/github.com/equinor/radix-networkpolicy-canary/main /app/radix-networkpolicy-canary

EXPOSE 5000

CMD ["/bin/sh", "-c", "/app/radix-networkpolicy-canary"]
