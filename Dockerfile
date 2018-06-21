# Building:
# docker build --tag radix-canary-golang:latest -f Dockerfile .
# docker run -p 5000:5000 -e "LISTEN_PORT=5000" radix-canary-golang:latest

# Private ACR:
# docker tag radix-canary-golang:latest radixdev.azurecr.io/radix-canary-golang:latest
# docker push radixdev.azurecr.io/radix-canary-golang:latest

# Public Dockerhub:
# docker tag radix-canary-golang:latest stianovrevage/radix-canary-golang:latest
# docker push stianovrevage/radix-canary-golang:latest

# Application build stage
FROM golang:1.10-alpine3.7 as build

ENV GOPATH /go

COPY . /go/

WORKDIR /go/src/equinor.com/radix-canary-golang/cmd/

RUN apk add --no-cache git
RUN go get -t -v ./... # go get -d -v
RUN go build -v main.go

# Application run stage
FROM alpine:3.7

# Add bash if you need an interactive shell in the container, adds ~4MB to final image
# RUN apk add --no-cache bash

COPY --from=build /go/src/equinor.com/radix-canary-golang/cmd/main /app/radix-canary-golang

EXPOSE 5000

CMD ["/bin/sh", "-c", "/app/radix-canary-golang"]
