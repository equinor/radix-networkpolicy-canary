# Building:
# docker build --tag radix-canary-golang:latest -f Dockerfile .
# docker run -p 5000:5000 -e "LISTEN_PORT=5000" radix-canary-golang:latest

# Private ACR:
# docker tag radix-canary-golang:latest radixdev.azurecr.io/radix-canary-golang:latest
# docker push radixdev.azurecr.io/radix-canary-golang:latest

# Public Dockerhub:
# docker tag radix-canary-golang:latest stianovrevage/radix-canary-golang:latest
# docker push stianovrevage/radix-canary-golang:latest

# docker tag radix-canary-golang:latest stianovrevage/radix-canary-golang:0.1.7
# docker push stianovrevage/radix-canary-golang:0.1.7

# Application build stage
FROM golang:1.14-alpine3.11 as build

ENV GOPATH /go

COPY . /go/src/github.com/equinor/radix-canary-golang/

WORKDIR /go/src/github.com/equinor/radix-canary-golang/

RUN apk add --no-cache git
RUN go get -t -v ./... # go get -d -v
RUN go build -v cmd/main.go

# Application run stage
FROM alpine:latest

# Add bash if you need an interactive shell in the container, adds ~4MB to final image
# RUN apk add --no-cache bash

RUN addgroup user && adduser -D -G user 2000
USER 2000

COPY --from=build /go/src/github.com/equinor/radix-canary-golang/main /app/radix-canary-golang

EXPOSE 5000

CMD ["/bin/sh", "-c", "/app/radix-canary-golang"]
